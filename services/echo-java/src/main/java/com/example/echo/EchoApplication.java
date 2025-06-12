package com.example.echo;

import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.net.InetSocketAddress;
import java.nio.charset.StandardCharsets;
import java.util.concurrent.Executors;

import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpHandler;
import com.sun.net.httpserver.HttpServer;

public class EchoApplication {

    public static void main(String[] args) throws IOException {
        int port = 8083; // Matching Dockerfile and docker-compose.yml
        HttpServer server = HttpServer.create(new InetSocketAddress(port), 0);

        server.createContext("/", new EchoHandler());
        server.createContext("/health", new HealthHandler());

        // Using a fixed thread pool, adjust as needed
        server.setExecutor(Executors.newFixedThreadPool(10)); 
        server.start();
        System.out.println("[java] Server started on port " + port);
    }

    static class EchoHandler implements HttpHandler {
        @Override
        public void handle(HttpExchange exchange) throws IOException {
            String method = exchange.getRequestMethod();
            System.out.println("[java] Received request: " + method + " " + exchange.getRequestURI());
            
            if ("POST".equalsIgnoreCase(method)) {
                InputStream requestBodyStream = exchange.getRequestBody();
                byte[] bodyBytes = requestBodyStream.readAllBytes();
                requestBodyStream.close();

                exchange.getResponseHeaders().set("Content-Type", "application/json");
                exchange.sendResponseHeaders(200, bodyBytes.length);

                OutputStream responseBody = exchange.getResponseBody();
                responseBody.write(bodyBytes);
                responseBody.close();
                
                System.out.println("[java] Request processed successfully");
            } else {
                System.out.println("Echo: Method not allowed - " + method);
                // Method Not Allowed
                exchange.sendResponseHeaders(405, -1); 
            }
        }
    }

    static class HealthHandler implements HttpHandler {
        @Override
        public void handle(HttpExchange exchange) throws IOException {
            String method = exchange.getRequestMethod();
            System.out.println("[java] Health check: " + method + " " + exchange.getRequestURI());
            
            if ("GET".equalsIgnoreCase(method)) {
                String response = "{\"status\":\"ok\"}";
                byte[] responseBytes = response.getBytes(StandardCharsets.UTF_8);

                exchange.getResponseHeaders().set("Content-Type", "application/json");
                exchange.sendResponseHeaders(200, responseBytes.length);

                OutputStream responseBody = exchange.getResponseBody();
                responseBody.write(responseBytes);
                responseBody.close();
                
                System.out.println("[java] Health check response sent: " + response);
            } else {
                System.out.println("[java] Health: Method not allowed - " + method);
                // Method Not Allowed
                exchange.sendResponseHeaders(405, -1); 
            }
        }
    }
}
