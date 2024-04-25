package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

// Thiết lập kết nối với MySQL
var db *sql.DB

func connectToDB() {
	var err error
	//"root" là tên máy chủ
	//NHLinh@17082003 là password
	//"/caffee" là tên database trong mysql của bạn
	dsn := "root:NHLinh@17082003@tcp(localhost:3306)/caffee" // Thay bằng thông tin thực tế
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Không thể kết nối với MySQL:", err)
	}
}

// Khai báo cấu trúc cho dữ liệu trong bảng "menu"
type MenuItem struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Price       int    `json:"price"`
	PriceL      int    `json:"price_L"`
	Description string `json:"description"`
	Img         string `json:"img"`
}

// Hàm xử lý yêu cầu GET
func getMenuHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Có thiết bị đang xem dữ liệu ")
	// Truy vấn dữ liệu từ bảng "menu"
	rows, err := db.Query("SELECT id, name, price, price_L, description, img FROM menu")
	if err != nil {
		log.Println("Lỗi khi truy vấn dữ liệu:", err)
		http.Error(w, "Không thể truy vấn dữ liệu", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Biến chứa danh sách menu
	var menu []MenuItem
	// Duyệt qua các dòng dữ liệu
	for rows.Next() {
		var item MenuItem
		if err := rows.Scan(&item.ID, &item.Name, &item.Price, &item.PriceL, &item.Description, &item.Img); err != nil {
			log.Println("Lỗi khi đọc dữ liệu:", err)
			http.Error(w, "Lỗi khi đọc dữ liệu", http.StatusInternalServerError)
			return
		}
		menu = append(menu, item)
	}

	// Trả về kết quả dưới dạng JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(menu)
}

// Hàm xử lý yêu cầu POST
func addMenuHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Dữ liệu vừa được thêm vào ")
	if r.Method != "POST" {
		http.Error(w, "Chỉ hỗ trợ POST", http.StatusMethodNotAllowed)
		return
	}

	var item MenuItem
	err := json.NewDecoder(r.Body).Decode(&item) // Đọc dữ liệu từ yêu cầu POST
	if err != nil {
		log.Println("Lỗi khi giải mã JSON:", err)
		http.Error(w, "Dữ liệu không hợp lệ", http.StatusBadRequest)
		return
	}

	// Chèn mục mới vào bảng "menu"
	result, err := db.Exec("INSERT INTO menu (name, price, price_L, description, img) VALUES (?, ?, ?, ?, ?)", item.Name, item.Price, item.PriceL, item.Description, item.Img)
	if err != nil {
		log.Println("Lỗi khi chèn dữ liệu:", err)
		http.Error(w, "Không thể thêm mục mới", http.StatusInternalServerError)
		return
	}

	// Lấy ID của mục mới được thêm vào
	lastInsertID, err := result.LastInsertId()
	if err != nil {
		log.Println("Lỗi khi lấy ID:", err)
		http.Error(w, "Lỗi khi lấy ID của mục mới", http.StatusInternalServerError)
		return
	}

	item.ID = int(lastInsertID) // Gán ID cho mục mới

	// Trả về mục mới dưới dạng JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

// Hàm xử lý yêu cầu xóa dữ liệu theo ID
func deleteMenuHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Nhận được yêu cầu xóa") // In ra thông tin về yêu cầu DELETE

	// Lấy tham số ID từ URL
	query := r.URL.Query()
	idStr := query.Get("id")
	if idStr == "" {
		http.Error(w, "ID là bắt buộc", http.StatusBadRequest)
		return
	}

	// Chuyển đổi ID từ chuỗi thành số nguyên
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID không hợp lệ", http.StatusBadRequest)
		return
	}

	// Xóa mục có ID tương ứng
	result, err := db.Exec("DELETE FROM menu WHERE id = ?", id)
	if err != nil {
		log.Println("Lỗi khi xóa dữ liệu:", err)
		http.Error(w, "Không thể xóa dữ liệu", http.StatusInternalServerError)
		return
	}

	// Kiểm tra số lượng dòng bị ảnh hưởng
	affectedRows, err := result.RowsAffected()
	if err != nil || affectedRows == 0 {
		http.Error(w, "Không tìm thấy mục với ID này", http.StatusNotFound)
		return
	}

	// Xác nhận xóa thành công
	w.WriteHeader(http.StatusNoContent) // Trả về HTTP 204 No Content để xác nhận xóa thành công
}

// Hàm xử lý yêu cầu tìm kiếm theo ID hoặc tên (phần chuỗi)
func searchHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Nhận được yêu cầu tìm kiếm")

	query := r.URL.Query()
	idStr := query.Get("id")
	name := query.Get("name")

	// Tìm kiếm theo ID
	if idStr != "" {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "ID không hợp lệ", http.StatusBadRequest)
			return
		}

		var item MenuItem
		err = db.QueryRow("SELECT id, name, price, price_L, description, img FROM menu WHERE id = ?", id).Scan(&item.ID, &item.Name, &item.Price, &item.PriceL, &item.Description, &item.Img)
		if err == sql.ErrNoRows {
			http.Error(w, "Không tìm thấy mục với ID này", http.StatusNotFound)
			return
		} else if err != nil {
			log.Println("Lỗi khi truy vấn dữ liệu:", err)
			http.Error(w, "Không thể tìm dữ liệu", http.StatusInternalServerError)
			return
		}

		// Trả về kết quả dưới dạng JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(item)
		return
	}

	// Tìm kiếm theo tên (sử dụng LIKE để tìm chuỗi con)
	if name != "" {
		// Sử dụng '%' để tìm chuỗi con
		rows, err := db.Query("SELECT id, name, price, price_L, description, img FROM menu WHERE name LIKE ?", "%"+name+"%")
		if err != nil {
			log.Println("Lỗi khi truy vấn dữ liệu:", err)
			http.Error(w, "Không thể tìm dữ liệu", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var results []MenuItem
		for rows.Next() {
			var item MenuItem
			err := rows.Scan(&item.ID, &item.Name, &item.Price, &item.PriceL, &item.Description, &item.Img)
			if err != nil {
				log.Println("Lỗi khi đọc dữ liệu:", err)
				http.Error(w, "Lỗi khi đọc dữ liệu", http.StatusInternalServerError)
				return
			}
			results = append(results, item)
		}

		// Nếu không có kết quả, trả về lỗi
		if len(results) == 0 {
			http.Error(w, "Không tìm thấy mục với tên này", http.StatusNotFound)
			return
		}

		// Trả về kết quả dưới dạng JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
		return
	}

	http.Error(w, "ID hoặc tên là bắt buộc", http.StatusBadRequest)
}

// Hàm xử lý yêu cầu chỉnh sửa dữ liệu theo ID
func updateMenuByIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PATCH" && r.Method != "PUT" {
		http.Error(w, "Chỉ hỗ trợ PATCH hoặc PUT", http.StatusMethodNotAllowed)
		return
	}

	// Lấy tham số ID từ query string
	query := r.URL.Query()
	idStr := query.Get("id")
	if idStr == "" {
		http.Error(w, "ID là bắt buộc", http.StatusBadRequest)
		return
	}

	// Chuyển đổi ID từ chuỗi sang số nguyên
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID không hợp lệ", http.StatusBadRequest)
		return
	}

	// Đọc dữ liệu từ yêu cầu PATCH/PUT
	var updatedItem MenuItem
	err = json.NewDecoder(r.Body).Decode(&updatedItem)
	if err != nil {
		http.Error(w, "Dữ liệu không hợp lệ", http.StatusBadRequest)
		return
	}

	// Truy vấn SQL để cập nhật dữ liệu theo ID
	queryStr := "UPDATE menu SET name = ?, price = ?, price_L = ?, description = ?, img = ? WHERE id = ?"
	_, err = db.Exec(queryStr, updatedItem.Name, updatedItem.Price, updatedItem.PriceL, updatedItem.Description, updatedItem.Img, id)
	if err != nil {
		log.Println("Lỗi khi cập nhật dữ liệu:", err)
		http.Error(w, "Không thể cập nhật dữ liệu", http.StatusInternalServerError)
		return
	}

	fmt.Println("Dữ liệu đã được cập nhật cho ID:", id)

	// Xác nhận cập nhật thành công bằng HTTP 200 OK
	w.WriteHeader(http.StatusOK)
}

func main() {
	connectToDB()
	http.HandleFunc("/get", getMenuHandler)       //http://localhost:8081/get
	http.HandleFunc("/add", addMenuHandler)       //http://localhost:8081/add
	http.HandleFunc("/delete", deleteMenuHandler) //http://localhost:8081/delete?id=7
	http.HandleFunc("/search", searchHandler)     //http://localhost:8081/search?id=1
	//localhost:8081/search?name=Cốt Dừa Cốm Xanh
	http.HandleFunc("/update", updateMenuByIDHandler) //http://localhost:8081/update?id=5
	log.Println("Máy chủ đang chạy tại cổng 8081...")
	http.ListenAndServe(":8081", nil)
}
