package main

/*
Copyright (c) 2020 Daniel Gurney

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

import (
	"flag"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/gonutz/w32"
)

const version = "0.0.2"

func getVersions(dir string) map[string]string {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	versions := make(map[string]string, 0)
	for _, file := range files {
		name := strings.ToLower(file.Name())
		// We are only interested in dll, exe, mui or sys files. We silently ignore any problematic files.
		if filepath.Ext(name) == ".exe" || filepath.Ext(name) == ".dll" || filepath.Ext(name) == ".sys" || filepath.Ext(name) == ".mui" {
			size := w32.GetFileVersionInfoSize(dir + `\` + name)
			if size == 0 {
				continue
			}

			info := make([]byte, size)
			w32.GetFileVersionInfo(dir+`\`+name, info)

			translation, success := w32.VerQueryValueTranslations(info)
			if !success {
				continue
			}

			version, success := w32.VerQueryValueString(info, translation[0], "FileVersion")
			if !success {
				continue
			}
			versions[name] = version
		}
	}

	return versions
}

func main() {
	dir := flag.String("d", ".", "Directory to read.")
	name := flag.Bool("n", false, "Show names in output.")
	win := flag.Bool("w", false, "Show Windows build tags that could theoretically be full Windows builds.")
	winsd := flag.Bool("wsd", false, "Show build tags that use older source depot format.")
	v := flag.Bool("v", false, "Show version and exit.")
	flag.Parse()

	if *v {
		fmt.Printf("versionlist v%s by Daniel Gurney.\n", version)
		return
	}

	vers := getVersions(*dir)

	if *win && *winsd {
		fmt.Println("You can only specify one filter at a time.")
		flag.PrintDefaults()
		return
	}

	switch {
	default:
		rawVersions := make([]string, 0)
		for _, v := range vers {
			rawVersions = append(rawVersions, v)
		}

		encountered := map[string]bool{}
		versions := make([]string, 0)
		for v := range rawVersions {
			if !encountered[rawVersions[v]] {
				encountered[rawVersions[v]] = true
				versions = append(versions, rawVersions[v])
			}
		}

		sort.Strings(versions)
		for _, version := range versions {
			switch {
			default:
				fmt.Println(version)
			case *win:
				r, _ := regexp.Compile(`^(5|6|10).[0-5]{1}.[^0][\d]{3,4}.[\d]{1,5}[[:space:]]\([[:alpha:]\S]+.[\d]{6}-`)
				if r.MatchString(version) {
					fmt.Println(version)
				}
			case *winsd:
				r, _ := regexp.Compile(`^5.[0-1]{1,2}.[0-9]{4}.[0-9]{1,4} built by: (\w+) ?(at: (\d+)-(\d+))?`)
				if r.MatchString(version) {
					fmt.Println(version)
				}
			}
		}
	case *name:
		names := make([]string, 0)
		for n := range vers {
			names = append(names, n)
		}

		sort.Strings(names)
		for _, name := range names {
			fmt.Printf("%s: %s\n", name, vers[name])
		}
	}
}
