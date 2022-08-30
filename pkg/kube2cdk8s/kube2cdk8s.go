package kube2cdk8s

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pulumi/kube2pulumi/pkg/kube2pulumi"
	"github.com/smallcase/kube2cdk8s/util"
	"github.com/spf13/viper"
)

func Kube2CDK8S(filePath string) (string, error) {
	path, _, err := kube2pulumi.Kube2PulumiFile(filePath, "", "typescript")
	if err != nil {
		return "", err
	}

	input, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(input), "\n")

	for i, line := range lines {
		if strings.Contains(line, "import") {
			lines[i] = ``
		}
	}
	output := strings.Join(lines, "\n")

	viper.New()
	viper.AddConfigPath("/tmp")
	viper.SetConfigName(filepath.Base(filePath))
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return "", err
	}

	n := viper.GetString("metadata.name")
	k := viper.GetString("kind")

	name := fmt.Sprintf("new k8s.Kube%s(this, \"%s\", {", k, n)

	re := regexp.MustCompile("(?m)[\r\n]+^.*const.*$")
	res := re.ReplaceAllString(output, name)

	re2 := regexp.MustCompile("(?m)[\r\n]+^.*apiVersion.*$")
	loc := re2.FindStringIndex(res)
	res2 := res
	if loc != nil {
		res2 = res[:loc[0]] + res[loc[1]:]
	}

	re3 := regexp.MustCompile("(?m)[\r\n]+^.*kind.*$")
	loc = re3.FindStringIndex(res2)
	res3 := res2
	if loc != nil {
		res3 = res2[:loc[0]] + res2[loc[1]:]
	}

	defer os.Remove(path)

	return res3, nil
}

func Kube2CDK8SMultiple(filePath string) (string, error) {
	var result string

	input, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	m := strings.Split(string(input), "---")

	for _, v := range m {

		if v == "" {
			continue
		}
		f, err := util.CreateTempFile([]byte(v))
		if err != nil {
			return "", err
		}

		res, err := Kube2CDK8S(f.Name())
		if err != nil {
			return "", err
		}

		result += res
		result += "\n"

		err = os.Remove(f.Name())
		if err != nil {
			return "", err
		}
	}

	return result, nil
}
