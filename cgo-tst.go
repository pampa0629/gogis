package main

/*
#include <stdio.h>
#include  <stdlib.h>

void printint(int v) {
    printf("printint: %d\n", v);
}

void* memalloc(int count) {
	return malloc(count);
}

void writemem(void* ptr, int count) {
	char * pb = (char*)ptr;
	for(int i=0;i<count; i++) {
		if(pb[i] != 100) {
			pb[i] = i;
		}
	}
}

 static void myprint(char* s) {
   printf("%s\n", s);
 }



static void myprint2(char* s, int len) {
   int i = 0;
   for (; i < len; i++) {
     printf("%c", *(s+i));
   }
   printf("\n");
}

*/
import "C"
import (
	"fmt"
	"unsafe"
)

type StringStruct struct {
	str unsafe.Pointer
	len int
}

func mainc() {
	str := "Hello from stdio"
	ss := (*StringStruct)(unsafe.Pointer(&str))
	// fmt.Println(unsafe.Pointer(&str))

	fmt.Println(ss)

	c := (*C.char)(unsafe.Pointer(ss.str))
	C.myprint2(c, C.int(len(str)))
}

func mainc2() {
	v := 42
	// C.printint(C.int(v))
	// ptr := C.memalloc(C.int(v))
	data := make([]byte, v)
	data[9] = 100

	C.writemem(unsafe.Pointer(&data[0]), C.int(v))
	fmt.Println(data)

	// bts := C.GoBytes(unsafe.Pointer(ptr), C.int(v))
	// bts := C.GoBytes(ptr, C.int(v))
	// fmt.Println(bts)

	cs := C.CString("Hello from stdio")
	C.myprint(cs)
	C.free(unsafe.Pointer(cs)) // yoko注，去除这行将发生内存泄漏
}
