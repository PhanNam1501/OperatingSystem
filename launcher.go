package main

import (
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

func main() {
	// args := []string{"A", "B", "C", "D"}
	var cmds []*exec.Cmd
	// Khai báo slice chứa pointer đến các Cmd
	// Mục đích: lưu lại reference để sau này Wait() hoặc Kill() được

	// for _, a := range args {
	for i := 0; i < 2; i++ {
		cmd := exec.Command("./2.Introduction/mem")
		// Tạo 1 Cmd object, tương đương với gõ terminal:
		//   ./2.Introduction/cpu A
		// Chưa chạy, chỉ chuẩn bị thôi
		cmd.Stdout = os.Stdout
		// Gán stdout của child process → stdout của parent (terminal)
		// Nếu không có dòng này, child process in ra nhưng mình không thấy gì
		cmd.Stderr = os.Stderr
		// Tương tự, gán stderr của child → stderr của parent
		// Để nếu child có lỗi gì thì cũng hiện ra terminal
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: false}
		// SysProcAttr: cấu hình thuộc tính process ở mức system call
		// Setpgid: false → KHÔNG tạo process group mới cho child
		// Nghĩa là child nằm cùng process group với parent
		// → Khi Ctrl+C, OS gửi SIGINT cho cả group → child cũng nhận được
		// Nếu Setpgid: true → child tách group riêng → Ctrl+C không kill được nó
		err := cmd.Start()
		// Start() chạy process con ở background (non-blocking)
		// Khác với Run() là Run() = Start() + Wait() (blocking)
		// Ở đây dùng Start() để chạy 4 process SONG SONG cùng lúc
		if err != nil {
			log.Fatal(err)
		}
		// Nếu Start() lỗi (vd: file không tồn tại, không có quyền chạy)
		// → in lỗi ra stderr và thoát chương trình với exit code 1
		cmds = append(cmds, cmd)
		// Thêm cmd vào slice cmds để sau này dùng Wait() và Kill()
	}

	// Bắt Ctrl+C → kill tất cả children
	sigCh := make(chan os.Signal, 1)
	// Tạo buffered channel chứa được 1 Signal
	// Buffer = 1 để đảm bảo signal không bị mất
	// Nếu unbuffered (buffer = 0), signal có thể đến trước khi goroutine sẵn sàng nhận → mất signal
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	// Đăng ký lắng nghe 2 signal:
	// - SIGINT: khi người dùng nhấn Ctrl+C
	// - SIGTERM: khi dùng lệnh `kill <pid>` từ terminal khác
	// Khi nhận được signal → đẩy vào sigCh channel

	go func() {
		// Tạo goroutine (lightweight thread) chạy song song với main
		// Goroutine này ngồi chờ signal, không block main flow
		<-sigCh
		// Block tại đây, chờ cho đến khi nhận được signal từ channel
		// Khi user nhấn Ctrl+C → SIGINT được đẩy vào sigCh → dòng này unblock
		for _, cmd := range cmds {
			if cmd.Process != nil {
				cmd.Process.Kill()
			}
		}
		// Lặp qua tất cả child processes
		// cmd.Process != nil: kiểm tra process đã được Start() thành công chưa
		// Kill(): gửi SIGKILL cho child process → chết ngay lập tức
		// SIGKILL khác SIGTERM: SIGKILL không thể bị bắt hay ignore
		os.Exit(1)
		// Thoát parent process với exit code 1
		// Nếu không có dòng này, main() vẫn đang Wait() ở dưới
		// Tuy Wait() sẽ return khi child bị kill, nhưng os.Exit(1)
		// đảm bảo thoát ngay lập tức, sạch sẽ
	}()

	for _, cmd := range cmds {
		cmd.Wait()
	}
}
