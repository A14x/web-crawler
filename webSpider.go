package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
  "strings"
	"sync"
)



func main() {
	//setting up channels for passing urls between goroutines
	unprocessed_URLs := make(chan [2]string, 1000)
	//processed_URLs := make(chan [2]string, 1000)
	//get_URLs := make(chan string, 1000)
	bodies := make(chan [2]string, 1000)

	//set up WaitGroup
	var wg sync.WaitGroup
	wg.Add(1)                   //not proper FIX

	getRequest("https://www.example.com/", bodies)
	i:=0
	bodyParser(bodies, unprocessed_URLs, &i)

	//anonymous function goroutine for handling get requests that need to be made
	go func(){
		var url [2]string
		fmt.Println("Get goroutine start")
		for{
			for j:=0; j<5; j++{
				url = <-unprocessed_URLs //takes any urls from channel that need to be fetched from
				go getRequest(url[1], bodies)
			}
			url = <-unprocessed_URLs //takes any urls from channel that need to be fetched from
			getRequest(url[1], bodies)
			fmt.Println("Get done")
		}
	}()

	//anonymous function goroutine to find and return links in html receiving links on processed URLs channel
	go func(){
		fmt.Println("bodyParser goroutine start")
		i=0;
		for{
			//limits the number of goroutines of this type running at the moment
			for i<10{
				go bodyParser(bodies, unprocessed_URLs, &i)
				//i++
				fmt.Println(i)
			}
		}
	}()

	wg.Wait()
}

//takes bodies from body channel and parses them to extract links
func bodyParser(bodies chan [2]string, unprocessed_URLs chan [2]string, goroutineDone *int){
	body := <-bodies //[0] contains parentURL and [1] contains body strings
	*goroutineDone ++
	//finds the index of each href in the body
	links := indexLinks(body[1])
	fmt.Println(links)

	//Converts indexes to actual string URLs
	linkStrings := stringLinks(links, body[1], body[0])
	fmt.Println(linkStrings)

	lengthLinkStrings := len(linkStrings) //variable avoids unnecessary clock cycles doing len twice
	if lengthLinkStrings != 0 {

		//format into several {parentURl, childURL} slices and return on the unprocessed_URLs chan
		tmpSlice := [2]string{"", ""}
		for i := 0; i< lengthLinkStrings; i++ {
			tmpSlice = [2]string{body[0], linkStrings[i]}
			unprocessed_URLs <-tmpSlice
		}
	}
	*goroutineDone = *goroutineDone -1
}

//Takes the index list from indexLinks() and the html body from getRequest() and returns a list of strings
func stringLinks(linkIndex []int, body string, currentUrl string) []string {
	//length of "href='" is 6
	bodySlice := strings.Split(body, "") //split string into slice
	lenBody := len(bodySlice) //these two lengths will be reused for breaking for loops
	lenLinkIndex := len(linkIndex) // ^
	linkStrings := []string{} // the []string that will be returned

	// iterate through all link indexes
	for i:= 0; i < lenLinkIndex; i++ {
		//check to see if at final item in index
		if linkIndex[i] != -1 {
			//initialise vars
			tmp := []string{} //for constructing the linkString to be appended to linkStrings

			//iterate through body starting at link index[i] + 6 (to account for length of href=")
			for j := (linkIndex[i] + 6); j < lenBody; j++ {

				//check to see if reached delimiter of string " or '
				if (bodySlice[j] != "\"") && (bodySlice[j] != "'") {
					//add char to []string
					tmp = append(tmp, bodySlice[j])
					fmt.Println(tmp)

				} else {

					fmt.Println("full link found", j)
					//join the runes (chars) together to create the fully processed string
					finalString := strings.Join(tmp, "")

					//check for invalid links beggining with /
					if tmp[0] == "/" {
						finalString = currentUrl + finalString
					}

					linkStrings = append(linkStrings, finalString) 	// construct string and append to linkStrings

					j = lenBody // breaks for loop j
				}
			}
		} else {
			fmt.Println("All link strings have been found")
			i = lenLinkIndex //breaks for loop i
		}
	}

	fmt.Println(linkStrings)
	return linkStrings

}

//Finds each instance of "href=" in the html
func indexLinks(body string) []int {
  linkIndex := []int{} //creates an empty slice to store locations of links

  //Didn't want to repeat this code but it should work Clean later?
  //finds first occurance of link in current body                   //CHANGE TO LOOK FOR STANDARD URL FORMAT TO WORK ON OTHER FILE TYPES
  currentIndex := strings.Index(body, "href=")
  //adds to slice of indexes
  linkIndex = append(linkIndex, currentIndex)

	if linkIndex[0] != -1 {
		//remove everything prior to most recent index
	  body = removeLeft(body, currentIndex+1)
	}

  flag := true
  i := 0 //counter
  //whileloop to loop until all links have been found
  for flag == true {
    //checks to see if all are found as strings.Index returns -1 if none are found
    if linkIndex[i] != -1 {
      //finds first occurance of link in current body
      currentIndex := strings.Index(body, "href=")

			//remove everything prior to most recent index +1
      body = removeLeft(body, currentIndex+1)

      //adds to slice of indexes
      linkIndex = append(linkIndex, currentIndex)



    //else all href links found so end loop
    } else {
      fmt.Println("all links found") //for debugging
      flag = false
    }
    //increment counter
    i++
  }
	//process linkIndex slice to contain correct values for whole body rather than substring of body
	if linkIndex[0] != -1 {
		for j := 1; j < len(linkIndex); j++ {
			if linkIndex[j] != -1 {
				linkIndex[j] += linkIndex[j-1] + 1 // adds previous index and 1(because an extra char is removed in removeLeft())
			}else{
				//exit loop as all values processed except final -1
				j = len(linkIndex)
			}
		}
	}

  return linkIndex
}
//function to remove all characters to the left of an index in a string remember to input currentIndex +1 in normal use case for this project
func removeLeft(body string, currentIndex int) string {

	tmp0 := strings.Split(body, "") //create slice from string

  tmp1 := []string{} //empty slice of strings
  for j:= (currentIndex); j < (len(tmp0)); j++ {

    tmp1 = append(tmp1, tmp0[j])

  }
  body = strings.Join(tmp1, "")

  return body
}

//Sends a http/s GET request to receive HTML for website
func getRequest(url string, bodies chan [2]string) {
	fmt.Println("get request func start")
  //sending GET request
  res, err := http.Get(url)

  //error handling of get request
  if err == nil {
		//collect response
	  body, err := ioutil.ReadAll(res.Body)
	  res.Body.Close()

	  //error handling for ioUtil
	  if err != nil {
			fmt.Println("2")
	    log.Fatal(err)
	  }

	  //Convert byte slice (array of bytes) to string
	  bodyStr := string(body)
	  tmpSlice:= [2]string{url, bodyStr}
		bodies <-tmpSlice

  } else {
		//there was error in get request (usually wrong protocol/ missing http//) so return empty body
		bodyStr := " "
		tmpSlice:= [2]string{url, bodyStr}
		bodies <-tmpSlice
	}
	fmt.Println("get request done")
}
