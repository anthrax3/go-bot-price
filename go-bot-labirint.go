// go-bot-labirint
package main

import (
	"fmt"
	"os"
	"strings"
	"strconv"  
  	"io/ioutil"	
	"io"
	"net/http"
	"time"
	"golang.org/x/net/html"
	"github.com/ddo/pick"
	"golang.org/x/net/html/charset"
)


// структура книги
type dataBook struct {
	name  string // название книги
	autor string // автор
	year  int    // год издания
	kolpages int // кол-во стрниц
	ves  int   // вес книги
	price int // цена для всех (обычная)
	pricediscount int // цена со скидкой которая видна	
	url string  // ссылка на источник данных
}

func namebook(httpBody io.Reader) []string {
  links := make([]string, 0)
  page := html.NewTokenizer(httpBody)
  for {
    tokenType := page.Next()
    if tokenType == html.ErrorToken {
      return links
    }
    token := page.Token()
    if tokenType == html.StartTagToken && token.DataAtom.String() == "meta" {
      for _, attr := range token.Attr {
        if attr.Key == "content" {
          links = append(links, attr.Val)
        }
      }
    }
  }
}

//парсинг Автора, массы и кол-во страниц в книге
func parsedescribebook(s []string) dataBook {
	var b dataBook
	for i:=0;i<len(s);i++ {
		switch s[i] {
			case "Автор(ы)": b.autor=s[i+1]
			case "Масса": b.ves,_=strconv.Atoi(s[i+1])	
			case "Количество страниц": b.kolpages,_=strconv.Atoi(s[i+1])		 
		}
	}
	return b
}

//получение страницы из урла url
func gethtmlpage(url string) []byte {
	resp, err := http.Get(url)
    if err != nil {
        fmt.Println("HTTP error:", err)
		panic("HTTP error")        
    }

    defer resp.Body.Close()
    // вот здесь и начинается самое интересное
    utf8, err := charset.NewReader(resp.Body, resp.Header.Get("Content-Type"))
    if err != nil {
        fmt.Println("Encoding error:", err)
        panic("Encoding error")
    }
    // оп-па-ча, готово
	
//	fmt.Println(namebook(utf8))
	
    body, err := ioutil.ReadAll(utf8)
    if err != nil {
        fmt.Println("IO error:", err)
		panic("IO error")
    }
	return body
}

//----- разбор html страницы сайта Лабиринт
func parselabirintbook (shtml string) dataBook {		
	var book dataBook
	
	scena, _ := pick.PickText(&pick.Option{   // текст цены книги
		&shtml,
		"span",
		&pick.Attr{
			"itemprop",
			"price",
		},
	})

	scenaskidka, _ := pick.PickText(&pick.Option{   // текст цены книги
		&shtml,
		"span",
		&pick.Attr{
			"class",
			"buying-pricenew-val-number",
		},
	})		
	
	sauthor, _ := pick.PickText(&pick.Option{   // текст описания книги
		&shtml,
		"span",
		&pick.Attr{
			"itemtype",
			"http://schema.org/ItemList",
		},
	})

	stitle, _ := pick.PickText(&pick.Option{&shtml,"span",&pick.Attr{"itemprop","name"}})	
	book=parsedescribebook(sauthor)
	book.name=stitle[1]
	if len(scenaskidka)>0 {
		book.pricediscount,_=strconv.Atoi(scenaskidka[0])
	}
	vv := strings.Split(scena[0], " ")
	book.price,_ =strconv.Atoi(vv[1])
	return book
}

func printbook (book dataBook) {
	fmt.Println("Автор: ",book.autor)
	fmt.Println("Название книги: ",book.name)
	fmt.Println("Вес: ",book.ves)
	fmt.Println("Кол-во страниц: ",book.kolpages)
	fmt.Println("Цена: ",book.price)
	fmt.Println("Цена со скидкой: ",book.pricediscount)
	fmt.Println("Ссылка на книгу: ",book.url)
	return
}




// чтение файла с именем namefи возвращение содержимое файла, иначе текст ошибки
func readfiletxt(namef string) string {
	file, err := os.Open(namef)
	if err != nil {
		return "handle the error here"
	}
	defer file.Close()
	// get the file size
	stat, err := file.Stat()
	if err != nil {
		return "error here"
	}
	// read the file
	bs := make([]byte, stat.Size())
	_, err = file.Read(bs)
	if err != nil {
		return "error here"
	}
	return string(bs)
}

// чтение из текстового конфиг файла и возращает массив строк
func readcfgs(namef string) []string {
//	var res []string
	str := readfiletxt(namef)
	vv := strings.Split(str, "\n")		
//	for i:=0;i<len(vv);i++ {
//		if vv[i]!="" {
//			res=append(res,vv[i])
//		}
//	}
	return vv
}

//сохранение строки str в файл с именем namef
func savestrtofile(namef string, str string) error {
	file, err := os.Create(namef)
	if err != nil {
		// handle the error here
		return err
	}
	defer file.Close()

	file.WriteString(str)
	return err
}

func getbooklabirint(url string) dataBook {
	body:=gethtmlpage(url)	
	
	shtml := string(body)	

	book:=parselabirintbook(shtml)
	
	return book
}

//сохранить данные dataBook в файл 
// stitle:="Дата выгрузки;Автор;Название книги;Год издания;Кол-во стр.;Вес;Цена;Цена со скидкой"
func (db *dataBook) savetocsvfile(namef string) error {
	var fileflag bool = false
	if _, err := os.Stat(namef); os.IsNotExist(err) {
 	 // path/to/whatever does not exist
		fileflag=true
	}
	
	file, err := os.OpenFile(namef, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
		// handle the error here
		return err
	}
	defer file.Close()
	if fileflag { // если не существует файл
		stitle:="Дата выгрузки;Автор;Название книги;Год издания;Кол-во стр.;Вес;Цена;Цена со скидкой;Ссылка"+"\n"
		file.WriteString(stitle)
	}
    curdate := time.Now().String()
	str:=curdate+";"+db.autor+";"+db.name+";"+strconv.Itoa(db.year)+";"+strconv.Itoa(db.kolpages)+";"+strconv.Itoa(db.ves)+";"+strconv.Itoa(db.price)+";"+strconv.Itoa(db.pricediscount)+";"+db.url+"\n"
	file.WriteString(str)
	return err
	return err
}


func main() {
	//sdir:="books"
	namestore:="labirint"	
	namefurls:=namestore+"-url.cfg"
////	namefhtml:=namestore+"-page.html"
	
	fmt.Println("Start programm....!")
//	// получаем урлы из файлы
    list_urls:=readcfgs(namefurls)
	
	for i:=0;i<len(list_urls);i++{
		book:=getbooklabirint(list_urls[i])
		namef:=	namestore+".csv"
		book.url=list_urls[i]
		book.savetocsvfile(namef)
		printbook(book)
	}

	fmt.Println("The end....!")
}
