package main

import (
	"errors"
	"net/http"
	"strings"
	"path"

	"github.com/rs/zerolog/log"
)

func vhostFromHostname(host string) (string, error) {
	pieces := strings.Split(host, ".")

	// If there are no dots, or only one dot, there's no vhost
	if len(pieces) == 1 || len(pieces) == 2 || len(pieces) == 3 && pieces[0] == "www" {
		return "", errors.New("No vhost")
	}

	if strings.HasPrefix(host, "0.0.0.0") ||
	 strings.HasPrefix(host, "127.0.0.1") ||
	 strings.HasSuffix(host, "fly.dev") {
		return "", errors.New("No vhost")
	}
	// This totally fails for IP-based hostnames

	// Otherwise, return the leftmost component
	return pieces[0], nil
}

func vhostify(base http.Handler, f http.FileSystem) http.Handler {
	vhosts := detectVhosts(f)
	for path, _ := range vhosts {
		log.Debug().Msgf("vhost found: %s", path)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vhost, err := vhostFromHostname(r.Host)
		if err != nil {
			log.Debug().Msgf("no vhost in hostname, serving default")
			base.ServeHTTP(w, r)
			return
		}

		host, exists := vhosts[vhost]
		if exists {
			log.Debug().Msgf("vhost found: %s", vhost)
			host.handler.ServeHTTP(w, r)
			return
		}
		log.Debug().Msgf("no vhost matching: %s", vhost)
		http.NotFound(w, r)

		// Here we need to pick a
		// convention.
		// wtf.fff.red -> fff.red/labs/wtf
		// nottoday.fff.red -> fff.red/labs/nottoday
		// That will work. Need to ensure
		// that fly.io generates a wildcard
		// cert

		// I would rather avoid adding
		// configuration for specifying
		// the domain name. It can be
		// inferred from the request.

		// otherwise, remove the leftmost
		// part of the DNS <name> and serve
		// files from labs/<name>
		//   -> make sure to confine
		//      to http root
		//      http.Dir handles this, however, it *will* follow
		//      symlinks. Since this web server is so minimal,
		//      use http.Dir and assume good intent when it comes
		//      to symlinks, and that any "hidden" .dotfiles
		//      such as .git are meant to be shared.
		//
		// This will, for example, work with xip
		//  http://wtf.127.0.0.1.xip.io
	})
}

type VHost struct {
	prefix string
	handler http.Handler
}

func detectVhosts(fileSystem http.FileSystem) map[string]VHost {
	vhostRoot, err := fileSystem.Open(*vhostPrefix)
	if err != nil {
		log.Debug().Msgf("Cannot read vhost directory %v", err)
		return make(map[string]VHost)
	}
	vhostDirs, err := vhostRoot.Readdir(512)
	vhosts := make(map[string]VHost)
	vhostBase := path.Join(*basePath, *vhostPrefix)

	for _, dir := range vhostDirs {
		if dir.IsDir() {
			name := dir.Name()

			// TODO reject any names that aren't DNS safe
			if (strings.Contains(name, " ")) {
				log.Debug().Str("vhost", name).Msg("Skipping vhost, non-DNS safe name")
				continue
			}
			//log.Printf("%v", path.Join(*basePath, name))
			vhosts[name] = VHost{name, http.FileServer(http.Dir(path.Join(vhostBase, name)))}
			//handler := handleReq()
		}
	}

	return vhosts
}
