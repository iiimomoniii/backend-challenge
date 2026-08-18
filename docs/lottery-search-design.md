1. Requirement
ออกแบบระบบค้นหาสลากกินแบ่งที่รองรับข้อมูลมากกว่า 10 ล้านรายการ ค้นหาแบบ Wildcard Pattern และป้องกันการแจก Ticket ซ้ำให้ผู้ใช้หลายคนพร้อมกัน โดยเป็น Design Proposal เท่านั้น

2. System Requirements
- Lottery tickets มากกว่า 10 ล้านรายการ
- Ticket เป็นเลข 6 หลัก
- Search ด้วย wildcard pattern (รูปแบบการค้นหาที่ใช้ตัวอักษรพิเศษแทนค่าอะไรก็ได้บางส่วนครับ
ในโจทย์ Lottery Search ตัว wildcard คือ * ซึ่งหมายถึง เลขอะไรก็ได้ 1 หลัก)
- Return ticket ที่ match pattern
- รองรับ concurrent users
- Ticket เดียวกันไม่ควรถูก allocate ให้ user หลายคนพร้อมกัน

3. Proposed Solution Architecture

Client -> API Gateway ->  Lottery API Service   -> Search ----> Elasticsearch
                                                -> Reserve ---> MongoDB
                                                -> Purchase --> MongoDB
                                                -> Cache -----> Redis

4. Data Structures

Lottery API Service เป็น Service หลักที่รับผิดชอบ Business Logic ของระบบ 
- การค้นหาสลากตาม Wildcard Pattern, 
- การ Reserve สลาก 
- การ Purchase สลาก โดยควบคุมการเปลี่ยนสถานะของสลากตาม Business Rule

หน้าที่ของแต่ละ Component
API Gateway รับ Request จาก Client และทำหน้าที่ เช่น:

- Authentication
- Rate Limiting
- Routing
- Request Validation เบื้องต้น
- Lottery API Service เป็น Application Layer สำหรับจัดการ API

Lottery API Service 
    รับ Request
    Validate Request
    เรียก Search Service
    เรียก Allocation Service (Reserve, Purchase)
    แปลงผลลัพธ์เป็น HTTP Response


Data model
{
  "id": "uuid", // UUID สำหรับระบุ Ticket แต่ละใบ
  "draw_id": "uuid",  // อ้างอิงไปยัง Lottery Draw
  "number": "123456", // หมายเลขสลาก 6 หลัก
  "status": "AVAILABLE", // สถานะของ Ticket
  "owner_id": null, // User ที่ถือครอง Ticket ถ้ามี
  "created_at": "...", // วันที่สร้างข้อมูล
  "updated_at": "..." // วันที่แก้ไขข้อมูลล่าสุด
}
แยก LotteryDraw ออกมา
{
  "id": "uuid",  // UUID สำหรับระบุงวดแบบไม่ซ้ำกัน
  "draw_date": "2026-08-16", // วันที่ออกรางวัล
  "year": 2026, // ปีของงวด
  "month": 8, // เดือนของงวด
  "sequence": 2, // ลำดับงวดภายในเดือน
  "created_at": "..." // วันที่สร้างข้อมูล
}

LotteryDraw 1 --> N LotteryTicket

เช่น
2026
    January
        Draw 1 (01/01/2026)
        Draw 2 (16/01/2026)
 
    February
        Draw 1 (01/02/2026)
        Draw 2 (16/02/2026)

Ticket Status
จะมี 
- AVAILABLE สลากใบนี้ยังไม่มีคนจองหรือซื้อ สามารถค้นหาและเลือกซื้อได้
{
  "number": "123456",
  "status": "AVAILABLE",
  "owner_id": null
}

- RESERVED เมื่อ user เลือกสลากและกด Reserve ระบบจะเปลี่ยนจาก AVAILABLE → RESERVED 
ระบุว่า user-001 กำลังถือสิทธิ์ในการซื้อสลากใบนี้อยู่
{
  "number": "123456",
  "status": "RESERVED",
  "owner_id": "user-001"
  "reserved_at": "...",
  "reservation_expires_at": "..." กรณี user ต้องคืนสลากกรณียังไม่ได้ทำการจ่ายเงิน
}

reserved_at , reservation_expires_at
เพื่อป้องกันกรณี User จองแล้วไม่ซื้อ ระบบสามารถปล่อยกลับมาเป็น AVAILABLE เมื่อหมดเวลา

- SOLD สลากถูกขายแล้ว เมื่อ user ที่จองไว้ทำการซื้อสำเร็จ
ระบุว่า user-001 ซื้อสลากสำเร็จ
{
  "number": "123456",
  "status": "SOLD",
  "owner_id": "user-001"
}

flow ปกติ
[AVAILABLE] -> Reserve -> [RESERVED] -> Purchase -> [SOLD]

กรณี Reservation หมดอายุ
[RESERVED] -> Reservation Expired -> [AVAILABLE]

5. Wildcard Pattern Matching Algorithm
เพราะเราไม่ควรเอา Pattern ไปค้นหาแบบ Full Scan ทั้ง 10 ล้านรายการ เช่น
10,000,000 tickets -> ตรวจทีละใบ -> Match หรือไม่ Match วิธีนี้จะมีค่าใช้จ่ายสูงเมื่อมี Request เข้ามาพร้อมกันจำนวนมาก

แต่ถ้า Parse Pattern ก่อน 
1****5 ->  prefix = 1 suffix = 5 -> ใช้ Index ค้นหา Candidate -> ได้เฉพาะเลขที่เกี่ยวข้อง จะช่วยลดจำนวนข้อมูลที่ต้องนำมาตรวจสอบอย่างมาก

เลขสลากทุกใบมีความยาว 6 หลัก และ Search Pattern ก็ต้องมีความยาว 6 ตัวอักษร ประกอบด้วย
0-9 = ต้องตรงกับตัวเลขนั้น
* = ตัวเลขอะไรก็ได้ในตำแหน่งนั้น

Pattern	Parse	            ความหมาย
123***	prefix=123	        ขึ้นต้นด้วย 123
***456	suffix=456	        ลงท้ายด้วย 456
1****5	prefix=1, suffix=5	ขึ้นต้นด้วย 1 และลงท้ายด้วย 5
**34**	position[3..4]=34	หลักที่ 3–4 ต้องเป็น 34
******	no constraint	    ทุกเลข Match

6. Indexing Strategy

เนื่องจากระบบต้องรองรับข้อมูลสลากมากกว่า 10 ล้านรายการ การค้นหาโดยอ่านข้อมูลทั้งหมด (Full Collection/Table Scan) ทุกครั้งจะไม่เหมาะสม จึงออกแบบ Index เพื่อช่วยลดจำนวนข้อมูลที่ต้องนำมาตรวจสอบ

ระบบสามารถแบ่ง Index ออกเป็น 3 ส่วนหลัก:

- Prefix Index
  ใช้สำหรับ Pattern ที่มีตัวเลขอยู่ด้านหน้า เช่น
    123***
    12****
    1****   
    สามารถสร้าง Index บน number เพื่อให้ Database สามารถค้นหาเลขที่ขึ้นต้นด้วย Prefix ได้โดยไม่ต้อง Scan ทั้ง 10 ล้านรายการ

- Composite Index
    นอกจากเลขสลากแล้ว ระบบมีข้อมูลเกี่ยวกับ งวด (Draw) และ สถานะของ Ticket
    {
        "draw_id": "draw-2026-08-16",
        "number": "123456",
        "status": "AVAILABLE"
    }
    ในการค้นหาจริง เราไม่ควรค้นหา Ticket ของทุกงวดพร้อมกัน เช่น User ต้องการค้นหาเฉพาะงวดปัจจุบัน
        ใช้ (draw_id, number) กับ search สำหรับค้นหา Ticket ภายในงวดที่ต้องการ
        ใช้ (draw_id, number, status) กับ Reserve เพื่อหา Ticket ที่ยังสามารถ Reserve ได้ 
                                     กับ Purchase ใช้แนวคิดเดียวกัน แต่ตรวจสอบว่า Ticket ถูก Reserve โดย user ที่ถูกต้อง
        Database สามารถใช้ Index เพื่อจำกัดข้อมูลให้แคบลงก่อน
            [10M+ Tickets] -> draw_id -> [เฉพาะงวดที่ต้องการ] -> number prefix -> [123***] -> status -> [AVAILABLE Tickets]

Search Index
    ถ้าต้องรองรับ Pattern ที่ซับซ้อนจำนวนมาก การใช้ Database Index เพียงอย่างเดียวอาจไม่เพียงพอ จึงใช้ Search Index เช่น Elasticsearch หรือ OpenSearch เพื่อรองรับ Pattern Matching
    
    Search Index จะเก็บข้อมูลที่จำเป็นสำหรับการค้นหา เช่น:
        {
            "ticket_id": "uuid",
            "draw_id": "draw-001",
            "number": "123456"
        }

    จากข้อมูล number = 123456 

    ยึดตามตำแหน่งตัวเลข
    position_1 = 1
    position_2 = 2
    position_3 = 3
    position_4 = 4
    position_5 = 5
    position_6 = 6

    แล้วทำให้เป็น ทำให้ Pattern เช่น 1****5 จะได้ว่า position_1 = 1 AND position_6 = 5
    แล้วก็เอา position_1 = 1 AND position_6 = 5 ไปค้นหาเฉพาะ Ticket ที่ตรงทั้งสองเงื่อนไข

7. Ticket Distribution Strategy
อธิบายเพิ่มว่า Search กับ Reserve เป็นคนละขั้นตอน เพราะ Search แค่หา Ticket ที่ตรง Pattern 
แต่ Reserve เป็นขั้นตอนที่ป้องกันการแจก Ticket ซ้ำ

สมมติ User A และ User B ค้นหา Pattern เดียวกันพร้อมกัน

Search: 1****5  
    ถ้า User A Reserve 123455 สำเร็จก่อน
    เมื่อ User B พยายาม Reserve Ticket เดียวกันจะไม่ได้เพราะ สถานะเปลี่ยนเป็น RESERVED แล้ว

User A                                                 User B                                                                           
                        ->123455 
                        -> Parse Pattern 
                        -> Search Index 
                        -> Candidate Tickets 
                        -> Filter status = AVAILABLE 
                        -> Atomic Reserve 
 -> Success                                            -> Failed 
 -> status = RESERVED                                  -> Try next
                                                                
8. Concurrency Control    
ใช้ Atomic Update ในขั้นตอน Reservation โดยตรวจสอบ status = AVAILABLE พร้อมกับเปลี่ยนสถานะเป็น RESERVED ใน operation เดียวกัน เพื่อป้องกัน Race Condition เมื่อมีหลาย User พยายาม Reserve Ticket เดียวกันพร้อมกัน                                                                                       

9. Database & Storage Recommendation

Primary Database
Technology: MongoDB
Responsibility: Source of Truth
เหตุผล: Flexible document model, รองรับข้อมูลจำนวนมาก และรองรับ Atomic Update / Transaction สำหรับการจัดการ Reservation และ Purchase

Search Index
Technology: Search Index
Responsibility: Wildcard / Pattern Search
เหตุผล: รองรับการค้นหา Pattern จากข้อมูลจำนวนมากได้รวดเร็ว และลดการ Scan ข้อมูลทั้งหมดจาก MongoDB

Cache / Temporary Storage
Technology: Redis
Responsibility: Cache / Temporary Reservation
เหตุผล: ลด Load จาก MongoDB และ Search Index รวมถึงสามารถใช้จัดการข้อมูลชั่วคราว เช่น Reservation TTL ได้

10. Performance Analysis
Elasticsearch ครับ เพราะโจทย์เน้น Wildcard / Pattern Search กับข้อมูล 10M+ records
 และต้องการแยก Search workload ออกจาก MongoDB

ตัวอย่าง MongoDB:
{
  "id": "uuid-001",
  "draw_id": "draw-2026-08-16",
  "number": "123456",
  "status": "AVAILABLE",
  "owner_id": null,
  "created_at": "...",
  "updated_at": "..."
}
Elasticsearch อาจเก็บเฉพาะข้อมูลที่ใช้ Search
{
  "ticket_id": "uuid-001",
  "draw_id": "draw-2026-08-16",
  "number": "123456",
  "position_1": "1",
  "position_2": "2",
  "position_3": "3",
  "position_4": "4",
  "position_5": "5",
  "position_6": "6"
}

สร้าง Ticket

[Create Ticket] ->  [MongoDB] -> Sync -> [Elasticsearch]

                    123456              ticket_id = uuid-001
                    AVAILABLE           number = 123456
                                        position_1 = 1
                                        position_2 = 2
                                        ...
                                        position_6 = 6

User Search: 1****5 
API จะ Parse Pattern ก่อน จาก 1****5 -> position_1 = 1 position_6 = 5
จากนั้นส่ง Query ไป Elasticsearch  (position_1 = 1 AND position_6 = 5)
Elasticsearch ค้นจาก Index แล้วคืน ticket_id: uuid-001

11. Scalability & Trade-offs
ระบบสามารถ Scale แต่ละ Component แยกกันตาม Workload ได้

API Service
สามารถเพิ่มจำนวน Instance ได้ตามจำนวน Request
               -> API #1
Load Balancer  -> API #2
               -> API #3       


MongoDB
ทำ Replica Set คือการมี MongoDB หลาย Instance ที่เก็บข้อมูลชุดเดียวกัน

Primary -> Secondary
        -> Secondary

Primary รับ Write และเป็นตัวหลักในการทำงาน
Secondary เป็นสำเนาของข้อมูล
ถ้า Primary มีปัญหา ระบบสามารถหมุนเอา Secondary ขึ้นมาเป็น Primary ได้

12. End-to-End Flow
Search 
Client -> API Gateway ->  Lottery API ->  Parse Pattern ->  Elasticsearch -> Candidate Ticket IDs ->
MongoDB -> Return Matching Tickets

Reserve
client ->  Lottery API ->  MongoDB          -> RESERVED
                         status = AVAILABLE
                         Atomic Update
  
Purchase
client ->  Lottery API ->  MongoDB          -> SOLD
                         status = RESERVED
                         owner_id = current_user

Reservation Expiration
MongoDB -> Find RESERVED tickets               -> Atomic Update -> จาก RESERVED เป็น AVAILABLE -> Clear Owner 
        where reservation_expires_at < NOW()
                   
                 
            
                                

