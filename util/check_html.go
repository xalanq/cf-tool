package util

import (
	"errors"
	"golang.org/x/net/html"
	"strings"
)

func CheckHtml(username string,body []byte) (error) {
	bodyString := string(body)
	doc,err := html.Parse(strings.NewReader(bodyString))

	if err != nil {
		return err
	}

	res := getElementById(doc,"sidebar",false)
	if res == nil {
		return errors.New("incorrect username/password")
	}

	res = getElementByClass(res,"personal-sidebar",false)

	if res != nil && res.Data != "div" {
		return errors.New("incorrect username/password")
	}

	return nil
}

func FetchUsername(body []byte) string {
	doc,_ := html.Parse(strings.NewReader(string(body)))

	res := getElementById(doc,"sidebar",false)
	res  = getElementByClass(res,"personal-sidebar",false)
	res  = getElementByClass(res,"avatar",false)
	res  = getElementByClass(res,"rated-user",true)

	fetchedUsername := strings.Split(res.Attr[0].Val,"/")
	return fetchedUsername[2]
}

func getElementById(node *html.Node, key string,partial bool) *html.Node {
	return traverse(node,key,"id",partial)
}

func getElementByClass(node *html.Node,key string,partial bool) *html.Node {
	return traverse(node,key,"class",partial)
}

func traverse(node *html.Node, key,findBy string,partial bool) *html.Node {
	if findBy == "id" {
		if checkId(node,key) {
			return node
		}
	}

	if findBy == "class" {
		if checkClass(node,key,partial) {
			return node
		}
	}

	if node == nil {
		return nil
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		result := traverse(c, key, findBy,partial)
		if result != nil {
			return result
		}
	}

	return nil
}

func checkClass(node *html.Node, class string,partial bool) bool {
	if node == nil {
		return false
	}

	if node.Type == html.ElementNode {
		s, ok := GetAttribute(node, "class")
		if ok {
			if partial {
				return strings.Contains(s,class)
			} else {
				return s == class
			}
		}
	}
	return false
}

func checkId(node *html.Node, id string) bool {
	if node.Type == html.ElementNode {
		s, ok := GetAttribute(node, "id")
		if ok && s == id {
			return true
		}
	}
	return false
}

func GetAttribute(node *html.Node, key string) (string, bool) {
	for _, attr := range node.Attr {
		if attr.Key == key {
			return attr.Val, true
		}
	}
	return "", false
}
