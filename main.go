package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type Product struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Price         int      `json:"price"`
	OldPrice      *int     `json:"oldPrice"`
	Image         string   `json:"image"`
	Categories    []string `json:"categories"`
	Year          *int     `json:"year"`
	Material      *string  `json:"material"`
	BladeLengthCM *int     `json:"bladeLengthCm"`
	BladeWidthMM  *int     `json:"bladeWidthMm"`
	Description   *string  `json:"description"`
}

var db *sql.DB

func main() {

	// 🔥 production mode
	gin.SetMode(gin.ReleaseMode)

	// ===== DATABASE =====
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("❌ DATABASE_URL not set")
	}

	var err error
	db, err = sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("❌ DB connection failed:", err)
	}

	log.Println("✅ Connected to database")

	// ===== SERVER =====
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type"},
	}))

	r.GET("/api/products", getProducts)
	r.POST("/api/products", addProduct)
	r.DELETE("/api/products/:id", deleteProduct)

	// ===== PORT (IMPORTANT FOR RAILWAY) =====
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("🚀 Server running on port", port)
	r.Run(":" + port)
}

// ================= GET PRODUCTS =================
func getProducts(c *gin.Context) {

	query := `SELECT id, name, price, old_price, image, categories, year, material,
 blade_length_cm, blade_width_mm, description FROM products`

	rows, err := db.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var products []Product

	for rows.Next() {
		var p Product
		var oldPrice, year, bladeLen, bladeWid sql.NullInt64
		var material, desc sql.NullString
		var categories pq.StringArray

		err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Price,
			&oldPrice,
			&p.Image,
			&categories,
			&year,
			&material,
			&bladeLen,
			&bladeWid,
			&desc,
		)
		if err != nil {
			continue
		}

		p.Categories = categories

		if oldPrice.Valid {
			v := int(oldPrice.Int64)
			p.OldPrice = &v
		}
		if year.Valid {
			v := int(year.Int64)
			p.Year = &v
		}
		if material.Valid {
			p.Material = &material.String
		}
		if bladeLen.Valid {
			v := int(bladeLen.Int64)
			p.BladeLengthCM = &v
		}
		if bladeWid.Valid {
			v := int(bladeWid.Int64)
			p.BladeWidthMM = &v
		}
		if desc.Valid {
			p.Description = &desc.String
		}

		products = append(products, p)
	}

	c.JSON(http.StatusOK, products)
}

// ================= ADD PRODUCT =================
func addProduct(c *gin.Context) {
	var p Product

	if err := c.BindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := db.Exec(`
 INSERT INTO products 
 (id, name, price, old_price, image, categories, year, material, blade_length_cm, blade_width_mm, description)
 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
 `,
		p.ID,
		p.Name,
		p.Price,
		p.OldPrice,
		p.Image,
		pq.Array(p.Categories),
		p.Year,
		p.Material,
		p.BladeLengthCM,
		p.BladeWidthMM,
		p.Description,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "product added"})
}

// ================= DELETE PRODUCT =================
func deleteProduct(c *gin.Context) {
	id := c.Param("id")

	_, err := db.Exec(`DELETE FROM products WHERE id=$1`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
