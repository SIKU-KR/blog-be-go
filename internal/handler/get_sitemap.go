package handler

import (
	"bumsiku/internal/repository"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// GetSitemap은 블로그의 모든 게시물과 카테고리를 포함하는 동적 sitemap.xml을 생성합니다.
func GetSitemap(postRepo *repository.PostRepository, categoryRepo *repository.CategoryRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()

		// 기본 도메인 URL 설정
		domain := "https://bumsiku.kr"

		// 현재 시간 (마지막 수정 시간)
		now := time.Now().Format("2006-01-02")

		// 모든 게시물 가져오기
		postsInput := &repository.GetPostsInput{
			Page:     1,
			PageSize: 999, // 충분히 큰 수로 설정하여 모든 게시물을 가져옵니다
		}
		postsOutput, err := postRepo.GetPosts(ctx, postsInput)
		if err != nil {
			c.XML(http.StatusInternalServerError, gin.H{"error": "게시물을 가져오는 중 오류가 발생했습니다"})
			return
		}

		// 모든 카테고리 가져오기
		categories, err := categoryRepo.GetCategories(ctx)
		if err != nil {
			c.XML(http.StatusInternalServerError, gin.H{"error": "카테고리를 가져오는 중 오류가 발생했습니다"})
			return
		}

		// XML 헤더 설정
		c.Header("Content-Type", "application/xml")

		// sitemap 시작 태그
		c.String(http.StatusOK, `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`)

		// 메인 페이지 URL
		c.String(http.StatusOK, fmt.Sprintf(`
  <url>
    <loc>%s</loc>
    <lastmod>%s</lastmod>
    <changefreq>daily</changefreq>
    <priority>1.0</priority>
  </url>`, domain, now))

		// 카테고리 URL
		for _, category := range categories {
			c.String(http.StatusOK, fmt.Sprintf(`
  <url>
    <loc>%s/category/%s</loc>
    <lastmod>%s</lastmod>
    <changefreq>weekly</changefreq>
    <priority>0.8</priority>
  </url>`, domain, category.Category, now))
		}

		// 게시물 URL
		for _, post := range postsOutput.Posts {
			lastmod := post.UpdatedAt.Format("2006-01-02")
			c.String(http.StatusOK, fmt.Sprintf(`
  <url>
    <loc>%s/post/%s</loc>
    <lastmod>%s</lastmod>
    <changefreq>monthly</changefreq>
    <priority>0.6</priority>
  </url>`, domain, post.PostID, lastmod))
		}

		// sitemap 종료 태그
		c.String(http.StatusOK, `
</urlset>`)
	}
}
