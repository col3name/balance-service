import http from 'k6/http';
import {sleep} from 'k6';

export const options = {
    iterations: 1000,
    vus: 100
};
// const data = JSON.stringify({
//     "description": "string",
//     "from": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
//     "to": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
//     "amount": 0
// });


let validator = new RegExp("^[a-z0-9]{8}-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{12}$", "i");

function gen(count) {
    let out = "";
    for (let i = 0; i < count; i++) {
        out += (((1 + Math.random()) * 0x10000) | 0).toString(16).substring(1);
    }
    return out;
}

function Guid(guid) {
    if (!guid) throw new TypeError("Invalid argument; `value` has no value.");

    this.value = Guid.EMPTY;

    if (guid && guid instanceof Guid) {
        this.value = guid.toString();

    } else if (guid && Object.prototype.toString.call(guid) === "[object String]" && Guid.isGuid(guid)) {
        this.value = guid;
    }

    this.equals = function (other) {
        // Comparing string `value` against provided `guid` will auto-call
        // toString on `guid` for comparison
        return Guid.isGuid(other) && this.value === other;
    };

    this.isEmpty = function () {
        return this.value === Guid.EMPTY;
    };

    this.toString = function () {
        return this.value;
    };

    this.toJSON = function () {
        return this.value;
    };
}

Guid.EMPTY = "00000000-0000-0000-0000-000000000000";

Guid.isGuid = function (value) {
    return value && (value instanceof Guid || validator.test(value.toString()));
};

Guid.create = function () {
    return new Guid([gen(2), gen(1), gen(1), gen(1), gen(3)].join("-"));
};

Guid.raw = function () {
    return [gen(2), gen(1), gen(1), gen(1), gen(3)].join("-");
};

// The default exported function is gonna be picked up by k6 as the entry point for the test script. It will be executed repeatedly in "iterations" for the whole duration of the test.
export default function () {
    // Make a GET request to the target URL
    http.get('http://localhost:8000/api/v1/money/67f9ff8c-79ea-4f39-a86e-39fb1d9dfb92/transactions?cursor=MjAyMi0wMi0yMyAxNjo1Nzo0MC4zMDMzMSArMDAwMCBVVEMhMSF0cnVl&sort=1&order=1');

    // const uuid = Guid.raw();
    // console.log(uuid)
    // http.post('http://localhost:8000/api/v1/money/transfer', data, {
    //     headers: {
    //         'Idempotency-Key': '55b2bcd0-2d09-498d-ae62-907a82484753'
    //     }
    // })
    // Sleep for 1 second to simulate real-world usage
    sleep(1);
}
// 55b2bcd0-2d09-498d-ae62-907a82484753
// 8c0276bd-d013-1c66-148d-982919369df9


