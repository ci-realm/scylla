package main

import (
	"net/url"
	"os"

	jenkins "github.com/yosida95/golang-jenkins"
)

func startJenkinsJob(jobName string, args url.Values) error {
	auth := &jenkins.Auth{
		Username: os.Getenv("JENKINS_API_USER"),
		ApiToken: os.Getenv("JENKINS_API_TOKEN"),
	}
	j := jenkins.NewJenkins(auth, os.Getenv("JENKINS_API_BASE_URL")+"/")

	job, err := j.GetJob(jobName)
	if err != nil {
		return err
	}

	return j.Build(job, args)
}

// type configJson struct {
// 	Auths map[string]map[string]string `json:"auths"`
// }

// type credentialStore struct{}

// func (c credentialStore) Basic(*url.URL) (string, string) {
// 	configJsonRaw, err := ioutil.ReadFile("/home/manveru/.docker/config.json")
// 	fail(err)

// 	config := configJson{}

// 	err = json.NewDecoder(bytes.NewBuffer(configJsonRaw)).Decode(&config)
// 	fail(err)

// 	pp.Println(config)

// 	dec := base64.NewDecoder(base64.StdEncoding, bytes.NewBufferString(config.Auths["quay.dc.xing.com"]["auth"]))
// 	authBytes, err := ioutil.ReadAll(dec)
// 	fail(err)
// 	authParts := bytes.SplitN(authBytes, []byte{':'}, 2)
// 	pp.Println(authParts)

// 	user := string(authParts[0])
// 	pass := string(authParts[1])
// 	pp.Println(user, pass)
// 	return user, pass
// }

// func (c credentialStore) RefreshToken(*url.URL, string) string     { return "" }
// func (c credentialStore) SetRefreshToken(*url.URL, string, string) {}

// type authTransport struct{}

// func (a authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
// 	configJsonRaw, err := ioutil.ReadFile("/home/manveru/.docker/config.json")
// 	fail(err)

// 	config := configJson{}

// 	err = json.NewDecoder(bytes.NewBuffer(configJsonRaw)).Decode(&config)
// 	fail(err)

// 	req.Header.Set("Authorization", "Bearer "+config.Auths["quay.dc.xing.com"]["auth"])
// 	pp.Println(req)
// 	return http.DefaultClient.Do(req)
// }

// func init() {
// 	ctx := context.Background()
// 	const baseURL = "https://quay.dc.xing.com/"
// 	const digestHash = "36cd5a1c7e6fba7e20a879d932d0ffac2e48beae7f153a368470006f591f3def"

// 	// cli, err := client.NewEnvClient()
// 	// if err != nil {
// 	// 	logger.Fatal(err)
// 	// }

// 	// name, err := reference.ParseNamed("quay.dc.xing.com/e-recruiting-api-team")
// 	// fail(err)

// 	// auth.NewAuthorizer(manager challenge.Manager, handlers ...auth.AuthenticationHandler)
// 	// manager := challenge.NewSimpleManager()
// 	// authHandler := auth.NewTokenHandler(nil, manager, nil, "quay.dc.xing.com", "push", "stat")
// 	// authorizer := auth.NewAuthorizer(manager, authHandler)
// 	// transport := Transport{}
// 	tr := transport.NewTransport(authTransport{})

// 	repo, err := client.NewRepository(ctx, name, baseURL, tr)
// 	if err != nil {
// 		logger.Fatal(err)
// 	}
// 	blobs := repo.Blobs(ctx)
// 	desc, err := blobs.Stat(ctx, digest.NewDigestFromHex("sha256", digestHash))
// 	if err != nil {
// 		logger.Fatal(err)
// 	}
// 	pp.Println(desc)

// 	// registry, err := client.NewRegistry(ctx, "https://quay.dc.xing.com/", http.DefaultTransport)
// 	// registry, err := client.NewRegistry("https://quay.dc.xing.com/", http.DefaultTransport)

// 	// fail(err)
// 	// named, err := reference.ParseNamed("quay.dc.xing.com/e-recruiting-api-team")
// 	// pp.Println(named)
// 	// fail(err)
// 	// ref, err := reference.ParseAnyReference("36cd5a1c7e6fba7e20a879d932d0ffac2e48beae7f153a368470006f591f3def")
// 	// fail(err)
// 	// pp.Println(ref)
// 	// repo, err := client.NewRepository(named, "https://quay.dc.xing.com/", http.DefaultTransport)
// 	// fail(err)
// 	// pp.Println(repo)
// 	// blobs := repo.Blobs(ctx)
// 	// blobs.Stat(ctx, digest.Digest("36cd5a1c7e6fba7e20a879d932d0ffac2e48beae7f153a368470006f591f3def"))

// 	// refdig, err := reference.WithDigest(named, digest.NewDigestFromHex("sha256", "36cd5a1c7e6fba7e20a879d932d0ffac2e48beae7f153a368470006f591f3def"))
// 	// fail(err)

// 	// hub, err := registry.New("https://quay.dc.xing.com/", user, pass)
// 	// fail(err)
// 	// pp.Println(hub.Ping())

// 	// tags, err := hub.Tags("e-recruiting-api-team/scylla")
// 	// pp.Println(tags, err)

// 	// readImage(registry, "result")

// 	// digest := digest.NewDigestFromHex("sha256", sha256)

// 	// hub.UploadBlob("e-recruiting-api-team/scylla", digest, stream)

// 	// tags, err := hub.Tags("e-recruiting-api-team/scylla")
// 	// fail(err)
// 	// pp.Println(tags)

// 	// repositories, err := hub.Repositories()
// 	// fail(err)

// 	// pp.Println(repositories)

// 	os.Exit(0)
// }

// func readImage(path string) {
// 	container, err := ioutil.ReadFile(path)
// 	fail(err)

// 	zr, err := gzip.NewReader(bytes.NewBuffer(container))
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	fmt.Printf("Name: %s\nComment: %s\nModTime: %s\n\n", zr.Name, zr.Comment, zr.ModTime.UTC())

// 	tr := tar.NewReader(zr)
// 	for {
// 		hdr, err := tr.Next()
// 		if err == io.EOF {
// 			break
// 		}
// 		fail(err)
// 		if hdr.Name == "./" {
// 			continue
// 		}
// 		if hdr.Name == "manifest.json" {
// 			manifests := parseManifest(tr)
// 			pp.Println(manifests)
// 		} else {
// 			if hdr.FileInfo().IsDir() {
// 				base := filepath.Base(hdr.Name)
// 				// hasBlob, err := hub.HasBlob("e-recruiting-api-team/scylla", digest.NewDigestFromHex("sha256", base))
// 				// fail(err)
// 				pp.Println(base)
// 				return
// 			}
// 		}
// 		fmt.Printf("Contents of %s:\n", hdr.Name)
// 	}

// 	if err := zr.Close(); err != nil {
// 		log.Fatal(err)
// 	}
// }

// // directories
// // Contents of fa7a3b1a25ae752ff0b4810d76080e759de5977fe4f85994a1b686fd0f7d4895/:
// // Contents of fa7a3b1a25ae752ff0b4810d76080e759de5977fe4f85994a1b686fd0f7d4895/VERSION:
// // Contents of fa7a3b1a25ae752ff0b4810d76080e759de5977fe4f85994a1b686fd0f7d4895/json:
// // Contents of fa7a3b1a25ae752ff0b4810d76080e759de5977fe4f85994a1b686fd0f7d4895/layer.tar:

// // manifest.json
// // [
// //   {
// //     "RepoTags": [
// //       "quay.dc.xing.com/e-recruiting-api-team/scylla:c897c9e08d7406e6bf51da5e265b6fa30fe4ffee"
// //     ],
// //     "Layers": [
// //       "c11a8204055dd9879e16155cff6992848825628bbec3aeb966a83db8ac43daad/layer.tar",
// //       ...
// //       "794549ac836afb15f4b434c486a472f63d3b3db94d84768c839648b0b9ddce05/layer.tar"
// //     ],
// //     "Config": "39dcf614b9bc56695f29cfe3bc4be68c0492c58657a0c09943955ac970c762ef.json"
// //   }
// // ]

// // /repositories
// // {
// // 	"quay.dc.xing.com\/e-recruiting-api-team\/scylla": {
// // 		"c897c9e08d7406e6bf51da5e265b6fa30fe4ffee": "c11a8204055dd9879e16155cff6992848825628bbec3aeb966a83db8ac43daad"
// // 	}
// // }

// type Manifest struct {
// 	RepoTags []string
// 	Layers   []string
// 	Config   string
// }

// func parseManifest(tr *tar.Reader) []Manifest {
// 	manifests := []Manifest{}
// 	err := json.NewDecoder(tr).Decode(&manifests)
// 	fail(err)
// 	return manifests
// }

// func fail(err error) {
// 	if err == nil {
// 		return
// 	}
// 	logger.Fatal(err)
// }
