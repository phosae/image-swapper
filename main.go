package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"

	"gomodules.xyz/jsonpatch/v2"
)

const (
	tlsKeyName  = "tls.key"
	tlsCertName = "tls.crt"
)

var conf Config

type Config struct {
	Registry map[string]string `json:"registry2registry"`
}

func main() {
	confPath := path.Join(os.Getenv("CONFIG_DIR"), "config.json")
	if b, err := os.ReadFile(confPath); err != nil {
		log.Fatalln("err parse config", err)
	} else {
		err := json.Unmarshal(b, &conf)
		if err != nil {
			log.Fatalln("err Unmarshal config.json", err)
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/mutate", mutate)
	certDir := os.Getenv("CERT_DIR")
	log.Println("serving https on 0.0.0.0:8000")
	log.Fatal(http.ListenAndServeTLS(":8000", filepath.Join(certDir, tlsCertName), filepath.Join(certDir, tlsKeyName), mux))
}

func mutate(w http.ResponseWriter, r *http.Request) {
	var (
		reviewReq, reviewResp admissionv1.AdmissionReview
		pd                    corev1.Pod
	)

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&reviewReq); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get pod object from request
	if err := json.Unmarshal(reviewReq.Request.Object.Raw, &pd); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	reviewResp.TypeMeta = reviewReq.TypeMeta
	reviewResp.Response = &admissionv1.AdmissionResponse{
		UID:     reviewReq.Request.UID, // write the unique identifier back
		Allowed: true,
		Result:  nil,
	}

	swapped := false
	for i := range pd.Spec.Containers {
		for origReg, targetReg := range conf.Registry {
			if strings.HasPrefix(pd.Spec.Containers[i].Image, origReg) {
				old := pd.Spec.Containers[i].Image
				pd.Spec.Containers[i].Image = strings.Replace(pd.Spec.Containers[i].Image, origReg, targetReg, 1)
				fmt.Printf("swap Pod %s/%s container=%s, image: %s => %s\n", pd.Namespace, pd.Name, pd.Spec.Containers[i].Name, old, pd.Spec.Containers[i].Image)
				swapped = true
				break
			}
		}
	}

	if swapped {
		pdJSON, err := json.Marshal(pd)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		patches, err := jsonpatch.CreatePatch(reviewReq.Request.Object.Raw, pdJSON)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		patchesJSON, err := json.Marshal(patches)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		reviewResp.Response.Patch = patchesJSON
		pt := admissionv1.PatchTypeJSONPatch
		reviewResp.Response.PatchType = &pt
	}

	returnJSON(w, reviewResp)
}

// returnJSON renders 'v' as JSON and writes it as a response into w.
func returnJSON(w http.ResponseWriter, v interface{}) {
	js, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}
