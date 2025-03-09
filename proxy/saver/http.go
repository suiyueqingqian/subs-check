package saver

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/bestruirui/bestsub/config"
	"github.com/bestruirui/bestsub/utils/log"
)

var (
	httpData     = make(map[string][]byte)
	httpDataLock sync.RWMutex
	httpServer   *http.Server
)

func getLocalIPs() []string {
	var ips []string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ips
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP.String())
			}
		}
	}
	return ips
}

func startHTTPServer() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		httpDataLock.RLock()
		defer httpDataLock.RUnlock()

		key := r.URL.Path[1:]
		if data, exists := httpData[key]; exists {
			w.Header().Set("Content-Type", "text/yaml; charset=utf-8")
			w.Header().Set("status", "ok")
			if _, err := w.Write(data); err != nil {
				http.Error(w, "Failed to write response", http.StatusInternalServerError)
			}
		} else {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>Please wait for the check to finish</title>
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body {
            font-family: Arial, sans-serif;
            background-color: #f5f5f5;
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100vh;
            margin: 0;
            padding: 20px;
            box-sizing: border-box;
        }
        .container {
            text-align: center;
            padding: 20px;
            background-color: white;
            border-radius: 12px;
            box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
            width: 100%%;
            max-width: 500px;
        }
        h1 {
            color: #333;
            margin-bottom: 20px;
            font-size: 24px;
        }
        p {
            color: #666;
            font-size: 16px;
            margin-bottom: 20px;
        }
        .loader {
            border: 6px solid #f3f3f3;
            border-top: 6px solid #3498db;
            border-radius: 50%%;
            width: 50px;
            height: 50px;
            animation: spin 1s linear infinite;
            margin: 0 auto 20px;
        }
        @keyframes spin {
            0%% { transform: rotate(0deg); }
            100%% { transform: rotate(360deg); }
        }
        .links {
            margin-top: 20px;
        }
        .links a {
            display: block;
            margin: 10px 0;
            padding: 10px 20px;
            background-color: #3498db;
            color: white;
            text-decoration: none;
            border-radius: 6px;
            transition: background-color 0.3s ease;
        }
        .links a:hover {
            background-color: #2980b9;
        }
    </style>
    <script>
        let checkInterval;
        const urls = ['/all.yaml', '/openai.yaml','/netflix.yaml','/disney.yaml','/youtube.yaml'];
        const urlStatus = {};  

        function checkStatus() {
            const linksContainer = document.getElementById('links');

            urls.forEach(url => {
                
                if (urlStatus[url] === true) return;

                fetch(url)
                    .then(response => {
                        if (response.ok) {
                           
                            if (!document.querySelector('a[href="' + url + '"]')) {
                                const link = document.createElement('a');
                                link.href = url;
                                link.textContent = 'Download ' + url;
                                linksContainer.appendChild(link);
                            }
                            
                            urlStatus[url] = true;
                            console.log('URL ready:', url);
                        } else {
                            
                            urlStatus[url] = false;
                        }
                    })
                    .catch(error => {
                        console.error('Error checking status:', error);
                        urlStatus[url] = false; 
                    });
            });

            const allReady = urls.every(url => urlStatus[url] === true);
            
            if (allReady) {
                clearInterval(checkInterval);
                document.querySelector('.loader').style.display = 'none';  
                document.querySelector('h1').textContent = 'Success';
                document.querySelector('p').textContent = 'All files are ready';
            }
        }

        window.onload = () => {
            urls.forEach(url => {
                urlStatus[url] = undefined;
            });
            checkStatus(); 
            checkInterval = setInterval(checkStatus, 1000); 
        };
    </script>
</head>
<body>
    <div class="container">
        <h1>Checking...</h1>
        <p>Please wait for the check to finish</p>
        <div class="loader"></div>
        <div id="links" class="links"></div>
    </div>
</body>
</html>`)
		}
	})

	httpServer = &http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%d", config.GlobalConfig.Save.Port),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	ips := getLocalIPs()

	for _, ip := range ips {
		log.Info("http server started at http://%s:%d", ip, config.GlobalConfig.Save.Port)
	}

	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error("http server error: %v", err)
	}
}

func SaveToHTTP(yamldata []byte, filename string) error {
	httpDataLock.Lock()
	defer httpDataLock.Unlock()
	httpData[filename] = yamldata
	return nil
}

func ValiHTTPConfig() error {
	return nil
}

func StartHTTPServer() {
	go startHTTPServer()
}
