package com.example.simplerestserver;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

@SpringBootApplication
public class SimpleRestServerApplication {

    public static void main(String[] args) {
        SpringApplication.run(SimpleRestServerApplication.class, args);
    }

    @RestController
    public class MyController {
        @GetMapping("/endpoint1")
        public String endpoint1() {
            return "Hello from endpoint 1";
        }

        @GetMapping("/health")
        public ResponseEntity<?> health() {
            return new ResponseEntity<>(HttpStatus.OK);
        }

        @GetMapping("/endpoint2")
        public String endpoint2() {
            return "Hello from endpoint 2";
        }

        @GetMapping("/endpoint3")
        public String endpoint3() {
            return "Hello from endpoint 3";
        }
    }
}

