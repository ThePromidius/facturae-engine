Auditor's Compliance Matrix
===========================

This matrix provides a direct map between legal requirements, our functional implementation, and the physical source code as proof of compliance.

+--------------------------+----------------------------+-----------------------------------+------------------------------------------+
| Legal Basis              | Functional Requirement     | Implementation Status             | Technical Proof (Source)                 |
+==========================+============================+===================================+==========================================+
| :ref:`hac_1177_art7`     | :ref:`req_hash_chaining`   | Implemented                       | :src:`src/internal/chain/types.go`       |
+--------------------------+----------------------------+-----------------------------------+------------------------------------------+
| :ref:`hac_1177_art13`    | :ref:`req_hash_chaining`   | Implemented                       | :src:`src/internal/chain/types.go`       |
+--------------------------+----------------------------+-----------------------------------+------------------------------------------+
| :ref:`hac_1177_art14`    | :ref:`req_signing`         | Implemented (XAdES-BES)           | :src:`src/internal/signing/p12.go`       |
+--------------------------+----------------------------+-----------------------------------+------------------------------------------+
| :ref:`crea_crece_art12`  | Interoperability           | Scheduled Q3 2026                 | N/A                                      |
+--------------------------+----------------------------+-----------------------------------+------------------------------------------+
| :ref:`hac_1177_art9`     | Audit Trail                | Implemented                       | :src:`src/internal/legal/events.go`      |
+--------------------------+----------------------------+-----------------------------------+------------------------------------------+

.. note::
   The status is updated automatically during our CI/CD pipeline audits.
