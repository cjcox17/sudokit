package pdf

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

var (
	ErrPDFGenerationFailed = errors.New("PDF generation failed")
	ErrNotImplemented      = errors.New("PDF service not implemented")
	ErrContextCanceled     = errors.New("PDF generation context canceled")
)

type EmbeddedImage struct {
	ContentID   string
	Filename    string
	ContentType string
	Data        []byte
}

type Service struct {
	baseURL         string
	timeout         time.Duration
	printBackground bool
	marginTop       float64
	marginBottom    float64
	marginLeft      float64
	marginRight     float64
	paperWidth      float64
	paperHeight     float64
	landscape       bool
	scale           float64
}

type Option func(*Service)

func WithBaseURL(url string) Option {
	return func(s *Service) {
		s.baseURL = url
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(s *Service) {
		s.timeout = timeout
	}
}

func WithPrintBackground(enable bool) Option {
	return func(s *Service) {
		s.printBackground = enable
	}
}

func WithMargins(top, right, bottom, left float64) Option {
	return func(s *Service) {
		s.marginTop = top
		s.marginRight = right
		s.marginBottom = bottom
		s.marginLeft = left
	}
}

func WithPaperSize(width, height float64) Option {
	return func(s *Service) {
		s.paperWidth = width
		s.paperHeight = height
	}
}

func WithLandscape(landscape bool) Option {
	return func(s *Service) {
		s.landscape = landscape
	}
}

func WithScale(scale float64) Option {
	return func(s *Service) {
		s.scale = scale
	}
}

func NewService(baseURL string, opts ...Option) *Service {
	s := &Service{
		baseURL:         baseURL,
		timeout:         30 * time.Second,
		printBackground: true,
		marginTop:       0,
		marginBottom:    0,
		marginLeft:      0,
		marginRight:     0,
		paperWidth:      8.5,
		paperHeight:     11.0,
		landscape:       false,
		scale:           1.0,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *Service) GenerateFromHTML(ctx context.Context, html string, images []EmbeddedImage) ([]byte, error) {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	htmlPart, err := writer.CreateFormFile("files", "index.html")
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create html form file: %v", ErrPDFGenerationFailed, err)
	}
	if _, err := htmlPart.Write([]byte(html)); err != nil {
		return nil, fmt.Errorf("%w: failed to write html content: %v", ErrPDFGenerationFailed, err)
	}

	if err := writer.WriteField("printBackground", "true"); err != nil {
		return nil, fmt.Errorf("%w: failed to write printBackground field: %v", ErrPDFGenerationFailed, err)
	}

	if err := writer.WriteField("marginTop", fmt.Sprintf("%.2f", s.marginTop)); err != nil {
		return nil, fmt.Errorf("%w: failed to write marginTop field: %v", ErrPDFGenerationFailed, err)
	}

	if err := writer.WriteField("marginBottom", fmt.Sprintf("%.2f", s.marginBottom)); err != nil {
		return nil, fmt.Errorf("%w: failed to write marginBottom field: %v", ErrPDFGenerationFailed, err)
	}

	if err := writer.WriteField("marginLeft", fmt.Sprintf("%.2f", s.marginLeft)); err != nil {
		return nil, fmt.Errorf("%w: failed to write marginLeft field: %v", ErrPDFGenerationFailed, err)
	}

	if err := writer.WriteField("marginRight", fmt.Sprintf("%.2f", s.marginRight)); err != nil {
		return nil, fmt.Errorf("%w: failed to write marginRight field: %v", ErrPDFGenerationFailed, err)
	}

	if err := writer.WriteField("paperWidth", fmt.Sprintf("%.2f", s.paperWidth)); err != nil {
		return nil, fmt.Errorf("%w: failed to write paperWidth field: %v", ErrPDFGenerationFailed, err)
	}

	if err := writer.WriteField("paperHeight", fmt.Sprintf("%.2f", s.paperHeight)); err != nil {
		return nil, fmt.Errorf("%w: failed to write paperHeight field: %v", ErrPDFGenerationFailed, err)
	}

	if s.landscape {
		if err := writer.WriteField("landscape", "true"); err != nil {
			return nil, fmt.Errorf("%w: failed to write landscape field: %v", ErrPDFGenerationFailed, err)
		}
	}

	if s.scale != 1.0 {
		if err := writer.WriteField("scale", fmt.Sprintf("%.2f", s.scale)); err != nil {
			return nil, fmt.Errorf("%w: failed to write scale field: %v", ErrPDFGenerationFailed, err)
		}
	}

	for _, img := range images {
		part, err := writer.CreateFormFile("files", img.Filename)
		if err != nil {
			return nil, fmt.Errorf("%w: failed to create form file for image: %v", ErrPDFGenerationFailed, err)
		}
		if _, err := part.Write(img.Data); err != nil {
			return nil, fmt.Errorf("%w: failed to write image data: %v", ErrPDFGenerationFailed, err)
		}
	}

	writer.Close()

	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/forms/chromium/convert/html", body)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create request: %v", ErrPDFGenerationFailed, err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{
		Timeout: s.timeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		if ctx.Err() == context.Canceled {
			return nil, fmt.Errorf("%w: %v", ErrContextCanceled, err)
		}
		return nil, fmt.Errorf("%w: request failed: %v", ErrPDFGenerationFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: gotenberg returned status %d: %s", ErrPDFGenerationFailed, resp.StatusCode, string(body))
	}

	pdfBuf, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to read response body: %v", ErrPDFGenerationFailed, err)
	}

	if len(pdfBuf) == 0 {
		return nil, fmt.Errorf("%w: empty PDF generated", ErrPDFGenerationFailed)
	}

	return pdfBuf, nil
}

func (s *Service) GenerateFromURL(ctx context.Context, url string) ([]byte, error) {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	if err := writer.WriteField("url", url); err != nil {
		return nil, fmt.Errorf("%w: failed to write url field: %v", ErrPDFGenerationFailed, err)
	}

	if err := writer.WriteField("printBackground", "true"); err != nil {
		return nil, fmt.Errorf("%w: failed to write printBackground field: %v", ErrPDFGenerationFailed, err)
	}

	if err := writer.WriteField("marginTop", fmt.Sprintf("%.2f", s.marginTop)); err != nil {
		return nil, fmt.Errorf("%w: failed to write marginTop field: %v", ErrPDFGenerationFailed, err)
	}

	if err := writer.WriteField("marginBottom", fmt.Sprintf("%.2f", s.marginBottom)); err != nil {
		return nil, fmt.Errorf("%w: failed to write marginBottom field: %v", ErrPDFGenerationFailed, err)
	}

	if err := writer.WriteField("marginLeft", fmt.Sprintf("%.2f", s.marginLeft)); err != nil {
		return nil, fmt.Errorf("%w: failed to write marginLeft field: %v", ErrPDFGenerationFailed, err)
	}

	if err := writer.WriteField("marginRight", fmt.Sprintf("%.2f", s.marginRight)); err != nil {
		return nil, fmt.Errorf("%w: failed to write marginRight field: %v", ErrPDFGenerationFailed, err)
	}

	if err := writer.WriteField("paperWidth", fmt.Sprintf("%.2f", s.paperWidth)); err != nil {
		return nil, fmt.Errorf("%w: failed to write paperWidth field: %v", ErrPDFGenerationFailed, err)
	}

	if err := writer.WriteField("paperHeight", fmt.Sprintf("%.2f", s.paperHeight)); err != nil {
		return nil, fmt.Errorf("%w: failed to write paperHeight field: %v", ErrPDFGenerationFailed, err)
	}

	if s.landscape {
		if err := writer.WriteField("landscape", "true"); err != nil {
			return nil, fmt.Errorf("%w: failed to write landscape field: %v", ErrPDFGenerationFailed, err)
		}
	}

	if s.scale != 1.0 {
		if err := writer.WriteField("scale", fmt.Sprintf("%.2f", s.scale)); err != nil {
			return nil, fmt.Errorf("%w: failed to write scale field: %v", ErrPDFGenerationFailed, err)
		}
	}

	writer.Close()

	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/forms/chromium/convert/url", body)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create request: %v", ErrPDFGenerationFailed, err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{
		Timeout: s.timeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		if ctx.Err() == context.Canceled {
			return nil, fmt.Errorf("%w: %v", ErrContextCanceled, err)
		}
		return nil, fmt.Errorf("%w: request failed: %v", ErrPDFGenerationFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: gotenberg returned status %d: %s", ErrPDFGenerationFailed, resp.StatusCode, string(body))
	}

	pdfBuf, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to read response body: %v", ErrPDFGenerationFailed, err)
	}

	if len(pdfBuf) == 0 {
		return nil, fmt.Errorf("%w: empty PDF generated", ErrPDFGenerationFailed)
	}

	return pdfBuf, nil
}
