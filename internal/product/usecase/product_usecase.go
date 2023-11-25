package usecase

import (
	"context"
	"encoding/csv"
	"fmt"
	beegoContext "github.com/beego/beego/v2/server/web/context"
	"github.com/radyatamaa/scrap-brick-app/internal/domain"
	"github.com/radyatamaa/scrap-brick-app/pkg/zaplogger"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

type productUseCase struct {
	zapLogger                          zaplogger.Logger
	contextTimeout                     time.Duration
	pgProductRepository            domain.PgProductRepository
}


func NewProductUseCase(timeout time.Duration,
	pgProductRepository            domain.PgProductRepository,
	zapLogger zaplogger.Logger) domain.ProductUseCase {
	return &productUseCase{
		pgProductRepository:            pgProductRepository,
		contextTimeout:                     timeout,
		zapLogger:                          zapLogger,
	}
}

// internal func
func (p productUseCase) writeToCSV(writer *csv.Writer, product domain.Product) error {
	// Write product details to the CSV file
	err := writer.Write([]string{product.Name, product.Desc, product.Image, product.Price, product.Rating, product.Merchant})
	if err != nil {
		return err
	}

	return err
}

func (p productUseCase) visitURL(url string) ([]domain.Product, error) {
back :
	// configure the browser options

	caps := selenium.Capabilities{}
	caps.AddChrome(chrome.Capabilities{Args: []string{
		//"--headless", // comment out this line for testing
	}})

	// create a new remote client with the specified options
	driver, err := selenium.NewRemote(caps, "")
	if err != nil {
		if err.Error() == "invalid session id: invalid session id" {
			goto back
		}
		return nil, err
	}

	// visit the target page
	err = driver.Get(url)
	if err != nil {
		if err.Error() == "invalid session id: invalid session id" {
			goto back
		}
		return nil, err
	}

	// retrieve the page raw HTML as a string
	// and logging it

	html, err := driver.PageSource()
	if err != nil {
		if err.Error() == "invalid session id: invalid session id" {
			goto back
		}
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	// Extract the data you need from the HTML code

	result := make([]domain.Product,0)

	doc.Find(".css-13l3l78 .css-20kt3o").Each(func(i int, s *goquery.Selection) {
		name := s.Text()
		result = append(result,domain.Product{
			Name:     name,
		})
	})

	doc.Find(".css-13l3l78 .css-20kt3o").Each(func(i int, s *goquery.Selection) {
		desc := s.Text()
		result[i].Desc = desc
	})

	doc.Find(".css-13l3l78 .css-54k5sq").Each(func(i int, s *goquery.Selection) {
		image, exists := s.Attr("href")
		if exists {
			result[i].Image = image
		}
	})

	doc.Find(".css-13l3l78 .css-o5uqvq").Each(func(i int, s *goquery.Selection) {
		price := s.Text()
		result[i].Price = price
	})

	doc.Find(".css-13l3l78 .css-1riykrk").Each(func(i int, s *goquery.Selection) {
		var rating int
		s.Find(".css-177n1u3").Each(func(i int, s *goquery.Selection) {
			_, exists := s.Attr("src")
			if exists {
				rating++
			}
		})
		result[i].Rating = fmt.Sprint(rating)
	})

	doc.Find(".css-13l3l78 .css-vbihp9").Each(func(i int, s *goquery.Selection) {
		merchant := s.Text()
		result[i].Merchant = merchant
	})

	return result, nil
}

func (p productUseCase) scraper(url string, writer *csv.Writer) {
	var wg sync.WaitGroup

	// Start scraping with multiple threads
	for i := 1; i <= 10; i++ {
		wg.Add(1)
		go func(pageNum int) {
			defer wg.Done()

			result := make([]domain.Product,0)

			// Retry up to 3 times
			for retry := 0; retry < 3; retry++ {
				list, err := p.visitURL(fmt.Sprint(url, "?page=", pageNum))
				if list == nil {
					p.zapLogger.Errorf("Goroutine %d: Error on attempt %d: %v\n", pageNum, retry+1, err)
					time.Sleep(30 * time.Second) // Wait before retrying
					continue
				}

				// Success
				p.zapLogger.Infof("Goroutine %d: Success on attempt %d: List Count %d \n", pageNum, retry+1,len(list))
				result = list
				break // Break out of the retry loop on success
			}

			if len(result) > 0{
				for _, product := range result {
					// Save to PostgreSQL database
					_,err := p.pgProductRepository.Store(context.Background(),product)
					if err != nil {
						p.zapLogger.Errorf("Goroutine %d: Error Insert To DB : %s",pageNum,err.Error())
					}

					// Write to CSV file
					err = p.writeToCSV(writer, product)
					if err != nil {
						p.zapLogger.Errorf("Goroutine %d: Error Insert To DB : %s",pageNum,err.Error())
					}
				}
			}
		}(i)
	}

	// Wait for all threads to finish
	wg.Wait()
}
///////


func (p productUseCase) ScrapeProducts(beegoCtx *beegoContext.Context, maxCount int) error {
	_, cancel := context.WithTimeout(beegoCtx.Request.Context(), p.contextTimeout)
	defer cancel()

	// CSV file setup
	csvFile, err := os.Create("products.csv")
	if err != nil {
		fmt.Println(err)
	}
	defer csvFile.Close()

	// CSV writer setup
	writer := csv.NewWriter(csvFile)
	defer writer.Flush()

	// Write CSV header
	writer.Write([]string{"Name", "Description", "Image Link", "Price", "Rating", "Merchant"})

	// Define the URL to scrape
	url := "https://www.tokopedia.com/p/handphone-tablet/handphone"


	// Perform the scraping
	p.scraper(url, writer)

	return err
}

