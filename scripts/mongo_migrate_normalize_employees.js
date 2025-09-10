// Normalize legacy employee documents to canonical field names
// Usage: run with mongosh against the target Mongo instance

const dbName = 'hris';
const collName = 'employees';
const db = db.getSiblingDB(dbName);

print('Starting employee normalization migration on DB:', dbName);

const cursor = db[collName].find({
  $or: [
    { employeeid: { $exists: true } },
    { legalname: { $exists: true } },
    { hiredate: { $exists: true } },
    { preferredname: { $exists: true } }
  ]
});

let count = 0;
cursor.forEach(doc => {
  const update = {};
  const unset = {};
  if (doc.employeeid && !doc.employee_id) {
    update.employee_id = doc.employeeid;
    unset.employeeid = "";
  }
  if (doc.legalname && !doc.legal_name) {
    update.legal_name = doc.legalname;
    unset.legalname = "";
  }
  if (doc.hiredate && !doc.hire_date) {
    update.hire_date = doc.hiredate;
    unset.hiredate = "";
  }
  if (doc.preferredname !== undefined && doc.preferredname !== null && !doc.preferred_name) {
    update.preferred_name = doc.preferredname;
    unset.preferredname = "";
  }
  if (Object.keys(update).length === 0 && Object.keys(unset).length === 0) {
    return;
  }
  const ops = {};
  if (Object.keys(update).length) ops.$set = update;
  if (Object.keys(unset).length) ops.$unset = unset;
  const res = db[collName].updateOne({ _id: doc._id }, ops);
  if (res.matchedCount && res.modifiedCount) {
    count += 1;
    print('Migrated _id:', doc._id.valueOf());
  }
});

print('Created/updated documents:', count);

// Second pass: ensure every document has a canonical employee_id (generate if missing)
print('Ensuring all docs have employee_id (generating where missing)...');
const missingCursor = db[collName].find({ $or: [ { employee_id: { $exists: false } }, { employee_id: null } ] });
let genCount = 0;
missingCursor.forEach(doc => {
  // if there's a legacy employeeid, we would have set it earlier; re-check
  if (doc.employee_id && doc.employee_id !== null) return;
  if (doc.employeeid) {
    // set from legacy field if present
    db[collName].updateOne({ _id: doc._id }, { $set: { employee_id: doc.employeeid }, $unset: { employeeid: "" } });
    genCount += 1;
    return;
  }
  // generate a stable id using timestamp + tail of ObjectId
  const generated = 'emp-' + Date.now() + '-' + doc._id.toString().slice(-6);
  db[collName].updateOne({ _id: doc._id }, { $set: { employee_id: generated } });
  genCount += 1;
});
print('Generated/filled employee_id for documents:', genCount);

// Ensure an index on employee_id for fast lookups and uniqueness; use partial filter to avoid nulls
print('Creating (unique) index on employee_id (partial to avoid nulls)...');
try {
  db[collName].createIndex({ employee_id: 1 }, { unique: true, partialFilterExpression: { employee_id: { $exists: true, $ne: null } } });
  print('Index created or already present.');
} catch (e) {
  print('Index creation warning:', e);
}

print('Migration complete.');
