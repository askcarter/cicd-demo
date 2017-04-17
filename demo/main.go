package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

var (
	version = 1
	color   = 0
)

var colors = [...]string{"blue", "green", "pink", "yellow"}

func main() {
	/*
	   This tool needs to be run from the sample-app folder,
	   because that's where the git repo lives.
	*/

	branch := ""
	// flag.IntVar(&counter, "a", 0, "what action to take")
	interval := flag.Int("i", 60, "how many seconds in between actions")
	// flag.StringVar(&branch, "b", "dev-test-branch", "what branch to create")
	flag.Parse()

	//  Infinitely loop and run git commands.
	counter := 0
	for {
		i := counter % 4
		fmt.Printf("\n\n\n\titeration: %v, index: %v\n\n\n", counter, i)
		switch i {
		case 0: // create dev branch
			// take a git action to kick off jenkins
			go func() {

				col := colors[(color+1)%len(colors)]
				// if branch == "" {
				branch = fmt.Sprintf("dev-%v-%v", col, time.Now().Unix())
				// }

				execCmds([][]string{
					{"git", "checkout", "-b", branch},
				})

				makeCodeChange()

				execCmds([][]string{
					{"git", "add", "-A"},
					{"git", "commit", "-m", fmt.Sprintf("\"changing the color to %v\"", col)},
					{"git", "push", "origin", branch},
				})

				// // This shouldn't be necessary since I'm in the cluster.
				// $ kubectl proxy
				// $ curl http://localhost:8001/api/v1/proxy/namespaces/new-feature/services/gceme-frontend:80/
			}()

		case 1: // merge dev branch to canary
			go func() {
				execCmds([][]string{
					{"git", "checkout", "canary"},
					{"git", "merge", branch},
					{"git", "branch", "-d", branch},
					{"git", "push", "origin", "canary"},
				})
			}()
		case 2: // delete dev branch
			go func() {
				execCmds([][]string{
					{"git", "push", "origin", ":" + branch},
					{"kubectl", "delete", "ns", branch},
				})
			}()
		case 3: // merge canary to production
			go func() {
				execCmds([][]string{
					{"git", "checkout", "master"},
					{"git", "merge", "canary"},
					{"git", "push", "origin", "master"},
					// {"kubectl", "rollout", "status", "--namespace=production", "deployment", "gceme-frontend-production"},
				})
			}()
			// <-time.After(time.Second * time.Duration(180))
		}
		<-time.After(time.Second * time.Duration(*interval))
		counter++
	}
}

func makeCodeChange() {
	/*
	   Change the color that the frontend displays.
	       //snip
	       <div class="card blue">
	       <div class="card-content white-text">
	       <div class="card-title">Backend that serviced this request</div>
	       //snip

	       sed -i '' s/blue/yellow/g html.go
	*/
	c1 := colors[color%len(colors)]
	color++
	c2 := colors[color%len(colors)]
	execCmds([][]string{
		{"sed", "-i.bak", fmt.Sprintf("s/%v/%v/g", c1, c2), "html.go"},
	})

	// versionStr := fmt.Sprintf("%v.0.0.", version)

	/*
	   Increment version string.
	       sed -i '' s/1.0.0/2.0.0/g main.go

	   So
	       `const version string = "1.0.0"``
	   becomes
	       `const version string = "2.0.0"`
	*/

	v1 := fmt.Sprintf("%v.0.0", version)
	version++
	v2 := fmt.Sprintf("%v.0.0", version)
	execCmds([][]string{
		{"sed", "-i.bak", fmt.Sprintf("s/%v/%v/g", v1, v2), "main.go"},
	})
}

func execCmds(cmds [][]string) {
	for _, c := range cmds {
		cmd := exec.Command(c[0], c[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		s := strings.Join(c, " ")
		log.Printf("\tRunning cmd [%v].\n", s)
		err := cmd.Run()
		if err != nil {
			log.Fatalf("cmd '%v': %v", s, err)
		}
	}
}
