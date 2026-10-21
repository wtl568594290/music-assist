package services

import (
	"bytes"
	"fmt"
	"image/gif"
	"image/jpeg"
	"image/png"
	"log"
	"musicassist/tools"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/ZeroHawkeye/wordZero/pkg/document"
	"github.com/ZeroHawkeye/wordZero/pkg/style"
	"github.com/disintegration/imaging"
	"github.com/pkg/errors"
	"golang.org/x/image/bmp"
)

type documentService struct {
}

type GenerateParams struct {
	PageSize           string
	PageTop            float64
	PageBottom         float64
	PageLeft           float64
	PageRight          float64
	CatalogCols        int
	CatalogFontSize    int
	PageNumberFontSize int
	SortType           tools.SortType
	SplitByDir         bool
	RemoteUrl          string
}

type docSource struct {
	Imgs []tools.FileInfo
	Name string
}

const (
	pageNumberSpacePt     = 3
	pageNumberLineMult    = 1.5
	pageNumberCellPercent = 0.18

	imageParagraphAfter      = 8
	imageParagraphFontSizePt = 10.5
	imageParagraphLineMult   = 1.16

	heightRedundancy = 2

	catalogLineMult = 1.5
)

var downloadPath = tools.ProjectName + "-download"

// pageBottom不设置，因为要显示页数，使用ImageHeightRedundancy来替代它
var pageTop, pageBottom, pageLeft, pageRight float64 = 5, 3, 20, 5
var catalogFontSizePt, pageNumberFontSizePt int = 12, 16
var maxH, maxW, imgFootSpace, pageNumberFootSPace float64
var pageSize document.PageSize = document.PageSizeA4

func init() {
	document.SetGlobalLevel(document.LogLevelWarn)
}

func NewDocumentService() *documentService {
	return &documentService{}
}

func generateDocSources(sources []tools.Source, imgs []tools.FileInfo, splitByDir bool) []docSource {
	docSrcs := []docSource{}
	if splitByDir {
		for _, s := range sources {
			if s.IsDir {
				tarDirImgs := []tools.FileInfo{}
				j := 0
				for _, f := range imgs {
					if f.Dir == s.Path {
						tarDirImgs = append(tarDirImgs, f)
					} else {
						imgs[j] = f
						j++
					}
				}
				imgs = imgs[:j]
				docSrcs = append(docSrcs, docSource{tarDirImgs, filepath.Base(s.Path)})
			}
		}
	}
	if len(imgs) > 0 {
		docSrcs = append(docSrcs, docSource{imgs, ""})
	}
	return docSrcs
}

func createDoc(imgs []tools.FileInfo, catalogCols int, appendName string) (string, error) {
	// Create new document
	doc := document.New()
	doc.SetPageSize(pageSize)
	doc.SetPageMargins(pageTop, pageRight, pageBottom, pageLeft)

	createCatalog(doc, imgs, catalogCols)
	doc.AddPageBreak()
	createImagesDoc(doc, imgs)

	if appendName != "" {
		appendName = "_" + appendName
	}
	docFileName := filepath.Join(downloadPath, time.Now().Format("20060102_150405")+appendName+".docx")
	if err := doc.Save(docFileName); err != nil {
		return "", err
	}
	cwd, _ := os.Getwd()
	return filepath.Join(cwd, docFileName), nil
}
func createCatalog(doc *document.Document, files []tools.FileInfo, cols int) {
	titlePara := doc.AddParagraph("目录")
	titlePara.SetStyle(style.StyleTitle)
	datePara := doc.AddParagraph(time.Now().Format("2006-01-02"))
	datePara.SetStyle(style.StyleSubtitle)
	catalogTab, err := doc.AddTable(&document.TableConfig{
		Cols:  cols,
		Rows:  1,
		Width: int(ptToTwip(mmToPt(maxW))),
	})

	if err != nil {
		return
	}

	catalogTab.SetTableBorders(&document.TableBorderConfig{InsideV: &document.BorderConfig{}})
	rowStr := []string{}
	for k, v := range files {
		if k%cols == 0 {
			if len(rowStr) > 0 {
				catalogTab.AppendRow(rowStr)
			}
			rowStr = []string{}
		}
		rowStr = append(rowStr, fmt.Sprintf("%d. %s", k+1, v.BaseName))
	}
	if len(rowStr) > 0 {
		if len(rowStr) < cols {
			for i := len(rowStr); i < cols; i++ {
				rowStr = append(rowStr, "")
			}
		}
		catalogTab.AppendRow(rowStr)
	}
	catalogTab.DeleteRow(0)
	for i := 0; i < catalogTab.GetRowCount(); i++ {
		for j := range cols {
			catalogTab.SetCellFormat(i, j, &document.CellFormat{
				TextFormat: &document.TextFormat{
					FontSize: catalogFontSizePt,
				},
			})
		}
	}
}

func createImagesDoc(doc *document.Document, files []tools.FileInfo) {
	for idx, imgFile := range files {
		img, err := imaging.Open(imgFile.Path)
		if err != nil {
			continue
		}
		w, h := img.Bounds().Dx(), img.Bounds().Dy()
		if w > h {
			img = imaging.Rotate270(img)
		}
		w, h = img.Bounds().Dx(), img.Bounds().Dy()
		var buf bytes.Buffer
		switch strings.ToLower(imgFile.Ext) {
		case ".jpg", ".jpeg":
			err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90})

		case ".png":
			err = png.Encode(&buf, img)

		case ".gif":
			err = gif.Encode(&buf, img, nil)

		case ".bmp":
			err = bmp.Encode(&buf, img)

		default:
			continue
		}
		if err != nil {
			continue
		}

		imgSize := &document.ImageSize{KeepAspectRatio: true}
		if float64(w)/float64(h) > maxW/maxH {
			imgSize.Width = maxW
		} else {
			imgSize.Height = maxH
		}

		_, err = doc.AddImageFromData(buf.Bytes(), fmt.Sprintf("image_%d_%d", idx, time.Now().Nanosecond()), document.ImageFormatJPEG, w, h,
			&document.ImageConfig{Size: imgSize,
				Position:  document.ImagePositionInline,
				Alignment: document.AlignRight,
				WrapText:  document.ImageWrapTopAndBottom,
			})
		if err != nil {
			continue
		}
		rowWidth := ptToTwip(mmToPt(maxW))
		table, err := doc.AddTable(&document.TableConfig{Cols: 3, Rows: 1, Width: int(rowWidth), ColWidths: []int{int(rowWidth * pageNumberCellPercent), int(rowWidth * (1 - pageNumberCellPercent*2)), int(rowWidth * pageNumberCellPercent)}})
		if err != nil {
			continue
		}
		// 不显示表格边框
		table.SetTableBorders(&document.TableBorderConfig{})
		trueImgH := maxH
		if imgSize.Width > 0 && imgSize.Height == 0 {
			trueImgH = imgSize.Width * float64(h) / float64(w)
		}
		// 最后一行需要插入一个空行，防止在表格后面又生成一个空白页
		var spaceLineH float64 = 0
		if idx == len(files)-1 {
			spaceLineH = 2
		}
		lineH := mmToPt(pageNumberFootSPace+maxH-trueImgH) - float64(spaceLineH)

		table.SetRowHeight(0, &document.RowHeightConfig{
			Height: int(lineH),
			Rule:   document.RowHeightExact,
		})
		table.SetCellFormattedText(0, 2, fmt.Sprintf("- %d -", idx+1), &document.TextFormat{
			FontSize: pageNumberFontSizePt,
		})
		table.SetCellFormat(0, 2, &document.CellFormat{
			HorizontalAlign: document.CellAlignRight,
			VerticalAlign:   document.CellVAlignBottom,
		})
		table.SetCellFormattedText(0, 1, imgFile.BaseName, &document.TextFormat{
			FontSize: pageNumberFontSizePt,
		})
		table.SetCellFormat(0, 1, &document.CellFormat{
			HorizontalAlign: document.CellAlignCenter,
			VerticalAlign:   document.CellVAlignBottom,
		})
		if spaceLineH > 0 {
			snap := false
			spacePara := doc.AddFormattedParagraph(".", &document.TextFormat{
				FontSize: 1,
			})
			spacePara.SetParagraphFormat(&document.ParagraphFormatConfig{
				LineSpacing: 1,
				BeforePara:  0,
				AfterPara:   1,
				SnapToGrid:  &snap,
			})
		}
	}
}

func ptToMm(pt float64) float64 {
	return pt * 25.4 / 72.0
}

func mmToPt(mm float64) float64 {
	return mm * 72.0 / 25.4
}

func ptToTwip(pt float64) float64 {
	return pt * 20
}

func (ds *documentService) GenerateDoc(params *GenerateParams) ([]string, error) {
	docs := []string{}
	imgsPathMap, err := generateImagesMap()
	if err != nil {
		return docs, err
	}
	if len(imgsPathMap) == 0 {
		return docs, errors.New("No image")
	}

	pageTop, pageBottom, pageLeft, pageRight = params.PageTop, params.PageBottom, params.PageLeft, params.PageRight
	catalogFontSizePt, pageNumberFontSizePt = params.CatalogFontSize, params.PageNumberFontSize
	imgFootSpace = ptToMm(imageParagraphAfter) + ptToMm(imageParagraphFontSizePt)*imageParagraphLineMult
	pageNumberFootSPace = ptToMm(pageNumberSpacePt)*2 + ptToMm(float64(pageNumberFontSizePt))*pageNumberLineMult
	var pageH, pageW float64 = 297, 210
	switch params.PageSize {
	case string(document.PageSizeA3):
		pageH, pageW, pageSize = 420, 297, document.PageSizeA3
	case string(document.PageSizeA4):
		pageH, pageW, pageSize = 297, 210, document.PageSizeA4
	case string(document.PageSizeA5):
		pageH, pageW, pageSize = 210, 148, document.PageSizeA5
	}
	maxH = pageH - pageTop - pageBottom - imgFootSpace - pageNumberFootSPace - heightRedundancy
	maxW = pageW - pageLeft - pageRight

	sources := getCurSources()
	imgs := imgsPathMap.ToSlice(params.SortType)
	docSrcs := generateDocSources(sources, imgs, params.SplitByDir)
	for _, docSrc := range docSrcs {
		if doc, err := createDoc(docSrc.Imgs, params.CatalogCols, docSrc.Name); err == nil {
			docs = append(docs, doc)
		}
	}
	if len(docs) == 0 {
		return docs, errors.New("Create document failed")
	}
	syncRemote(params.RemoteUrl, docSrcs)
	return docs, nil
}

// OpenDoc 打开指定路径的文件，无黑框闪烁
func (ds *documentService) OpenDoc(path string) error {
	// 创建命令
	cmd := exec.Command("cmd", "/c", "start", "", path)

	// 设置SysProcAttr来隐藏控制台窗口
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true, // 关键：隐藏子进程的控制台窗口
	}

	// 启动命令
	return cmd.Start()
}

func (ds *documentService) RemoveFile(paths []string) error {
	var res error = nil
	for _, path := range paths {
		err := os.Remove(path)
		if err != nil {
			log.Println(err)
			res = err
		}
	}
	return res
}
