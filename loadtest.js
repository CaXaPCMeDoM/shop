import http from 'k6/http';
import {check, sleep} from 'k6';
import {Trend} from 'k6/metrics';

const BASE_URL = 'http://localhost:8080';
const TOKEN_TREND = new Trend('token_time');
const PURCHASE_TREND = new Trend('purchase_time');
const TRANSFER_TREND = new Trend('transfer_time');

export let options = {
    scenarios: {
        auth_users: {
            executor: 'ramping-vus',
            startVUs: 0,
            stages: [
                {duration: '10s', target: 2000},
            ],
        },
        purchase: {
            executor: 'constant-vus',
            vus: 1000,
            startTime: '10s',
            duration: '10s',
        },
        transfer: {
            executor: 'constant-vus',
            vus: 1000,
            startTime: '20s',
            duration: '10s',
        },
    },
};

let users = new Map();

export default function () {
    let userId = `user${__VU}`;
    let isAuthPhase = __ITER === 0 && __VU <= 2000;

    if (isAuthPhase) {
        let loginRes = http.post(`${BASE_URL}/api/auth`, JSON.stringify({
            username: userId,
            password: 'password',
        }), {headers: {'Content-Type': 'application/json'}});

        check(loginRes, {'Auth successful': (res) => res.status === 200});

        if (loginRes.status === 200) {
            let token = JSON.parse(loginRes.body).token;
            users.set(userId, token);
            TOKEN_TREND.add(loginRes.timings.duration);
        }
    }

    sleep(1);

    if (__VU <= 1000) {
        let item = 'pen';
        let purchaseRes = http.get(`${BASE_URL}/api/buy/${item}`, {
            headers: {Authorization: `Bearer ${users.get(userId) || ''}`},
        });

        check(purchaseRes, {'Purchase successful': (res) => res.status === 200});
        PURCHASE_TREND.add(purchaseRes.timings.duration);
    } else if (__VU > 1000 && __VU <= 2000) {
        let receiverId = `user${__VU - 1000}`;
        let transferRes = http.post(`${BASE_URL}/api/sendCoin`, JSON.stringify({
            toUser: receiverId,
            amount: Math.floor(Math.random() * 100),
        }), {
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${users.get(userId) || ''}`,
            },
        });

        check(transferRes, {'Transfer successful': (res) => res.status === 200});
        TRANSFER_TREND.add(transferRes.timings.duration);
    }
}
