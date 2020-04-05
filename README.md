# This is going to be a simple web spider/crawler to familiarise myself with the go language

# Aim
  - Given a seed URL
  - Send GET request or Header request for webpage.
  - Scan webpage for Regex or hrefs
  - scan linked web pages recursively
  - create a new go routine with each new webpage scan to maximise efficiency
  - have a cap on number of goroutines to avoid maxing out network
  - keep record of URLs searched so not to search same page twice
  - use rate limiting and proxy settings to avoid getting blocked `more research needed`




#Libraries
  - net/http  `http requests` https://golang.org/pkg/net/http/
  - fmt (for console logging)
  - html `allows formatting to get rid of escape chars e.g. "&lt;" -> "<"` https://golang.org/pkg/html/
  - web interface stuff https://golang.org/doc/articles/wiki/
