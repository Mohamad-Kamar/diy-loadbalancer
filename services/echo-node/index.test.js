const request = require('supertest');
const express = require('express');
const app = require('./index');

describe('Echo Node Service', () => {
  // Echo endpoint tests
  describe('POST /', () => {
    it('should echo simple POST body', async () => {
      const res = await request(app)
        .post('/')
        .send({ msg: 'hi' })
        .set('Content-Type', 'application/json');
      expect(res.statusCode).toBe(200);
      expect(res.body).toEqual({ msg: 'hi' });
    });

    it('should echo complex JSON object', async () => {
      const complexBody = {
        string: 'value',
        number: 42,
        boolean: true,
        array: [1, 2, 3],
        nested: {
          key: 'value',
          arr: ['a', 'b']
        }
      };
      const res = await request(app)
        .post('/')
        .send(complexBody)
        .set('Content-Type', 'application/json');
      expect(res.statusCode).toBe(200);
      expect(res.body).toEqual(complexBody);
    });

    it('should handle empty JSON object', async () => {
      const res = await request(app)
        .post('/')
        .send({})
        .set('Content-Type', 'application/json');
      expect(res.statusCode).toBe(200);
      expect(res.body).toEqual({});
    });

    it('should reject invalid JSON', async () => {
      const res = await request(app)
        .post('/')
        .set('Content-Type', 'application/json')
        .send('not json');
      expect(res.statusCode).toBe(400);
    });

    it('should reject non-JSON content-type', async () => {
      const res = await request(app)
        .post('/')
        .set('Content-Type', 'text/plain')
        .send('hello');
      expect(res.statusCode).toBe(415);
    });
  });

  // Health endpoint tests
  describe('GET /health', () => {
    it('should return health status', async () => {
      const res = await request(app).get('/health');
      expect(res.statusCode).toBe(200);
      expect(res.body).toEqual({ status: 'ok' });
    });

    it('should reject POST to health endpoint', async () => {
      const res = await request(app)
        .post('/health')
        .send({ status: 'ok' });
      expect(res.statusCode).toBe(405);
    });
  });

  // Method not allowed tests
  describe('Method Not Allowed', () => {
    const methods = ['put', 'patch', 'delete'];
    
    methods.forEach(method => {
      it(`should reject ${method.toUpperCase()} requests to root`, async () => {
        const res = await request(app)[method]('/');
        expect(res.statusCode).toBe(405);
      });
    });
  });

  // Headers tests
  describe('Response Headers', () => {
    it('should return correct content-type header', async () => {
      const res = await request(app)
        .post('/')
        .send({ test: true })
        .set('Content-Type', 'application/json');
      expect(res.headers['content-type']).toMatch(/application\/json/);
    });
  });
});
