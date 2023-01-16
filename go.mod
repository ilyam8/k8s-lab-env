module k8s-lab-env

go 1.14

replace k8s.io/client-go => k8s.io/client-go v0.18.3

require (
	github.com/encryptio/alias v0.0.0-20151210173825-4f70d72df1d4
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/sirupsen/logrus v1.9.0
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	golang.org/x/time v0.0.0-20200416051211-89c76fbcd5d1 // indirect
	k8s.io/api v0.26.0
	k8s.io/apimachinery v0.26.0
	k8s.io/client-go v11.0.0+incompatible
)
