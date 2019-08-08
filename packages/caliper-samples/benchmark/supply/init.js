/*
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

'use strict';

module.exports.info  = 'Creating product...';

let txIndex = 0;
let ID = ['1q2', '1w2', 'a12', 'dd4'];
let owners = ['Alice', 'Bob', 'Claire', 'David'];
let bc, contx;

module.exports.init = function(blockchain, context, args) {
    bc = blockchain;
    contx = context;

    return Promise.resolve();
};

module.exports.run = function() {
    txIndex++;
    let productName = 'product_' + txIndex.toString() + '_' + process.pid.toString();
    let productID = ID[txIndex % ID.length];
    let productType = (((txIndex % 10) + 1) * 10).toString(); // [10, 100]
    let productOwner = owners[txIndex % owners.length];

    let args;
    if (bc.bcType === 'fabric-ccp') {
        args = {
            chaincodeFunction: 'initProduct',
            chaincodeArguments: [productID, productName, productType, productOwner],
        };
    } else {
        args = {
            verb: 'initProduct',
            name: productName,
            id: productID,
            type: productType,
            owner: productOwner
        };
    }

    return bc.invokeSmartContract(contx, 'supply', 'v1', args, 30);
};

module.exports.end = function() {
    return Promise.resolve();
};