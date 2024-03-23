# get raw tx
import requests

tx_hash = "0adc86b59ef3329c3d85eaafbde3ef071c6030e3b58386980e3122a68f679eef"
api_response = requests.get(f'https://blockchain.info/rawtx/{tx_hash}?format=hex')
# print(api_response.text)

# compute Hash of tx
import hashlib
import requests

tx = api_response.text
# tx = 

print(tx)

tx_hash = hashlib.sha256(bytes.fromhex(tx)).hexdigest()
tx_hash = hashlib.sha256(bytes.fromhex(tx_hash)).hexdigest()
print(tx_hash)

# 0100000001218883460f07350f167517c909b16b240bb738bb8ba493b1594f9facb45cfac60100000000ffffffff02546000000000000016001437fff1c9ce1d770cf82b38a1cdeba3972cddbb08a2ee8a00000000001600147ef8d1162a3f3691023a6fccb7723edd126ac80a00000000
# 0100000001218883460f07350f167517c909b16b240bb738bb8ba493b1594f9facb45cfac60100000000ffffffff02546000000000000016001437fff1c9ce1d770cf82b38a1cdeba3972cddbb08a2ee8a00000000001600147ef8d1162a3f3691023a6fccb7723edd126ac80a00000000
# ccc6253c11d89af11bdcf20bdf513112ff3c0bc6fe526045542e1edf9ac6ecd6 // from data
# 0adc86b59ef3329c3d85eaafbde3ef071c6030e3b58386980e3122a68f679eef // 225 > inputs testing Compact size