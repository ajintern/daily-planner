const API_BASE = "http://localhost:8083/api";
let currentFilter = 'all'; // State variable for filtering

document.addEventListener('DOMContentLoaded', () => {
    loadTasks();
    
    // Set default dates in form
    const now = new Date();
    document.querySelector('input[name="date"]').valueAsNumber = now.getTime();
    document.querySelector('input[name="time"]').value = now.toTimeString().slice(0,5);

    // Handle Add Submit
    document.getElementById('addForm').addEventListener('submit', async (e) => {
        e.preventDefault();
        const formData = new FormData(e.target);
        
        try {
            await fetch(`${API_BASE}/todo/add`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    text: formData.get('task'),
                    date: formData.get('date'),
                    time: formData.get('time')
                })
            });
            e.target.reset(); 
            
            // Re-set defaults
            const newNow = new Date();
            document.querySelector('input[name="date"]').valueAsNumber = newNow.getTime();
            document.querySelector('input[name="time"]').value = newNow.toTimeString().slice(0,5);
            
            loadTasks();
        } catch (err) {
            console.error(err);
            alert("Failed to add task");
        }
    });

    // Handle Edit Submit
    document.getElementById('editForm').addEventListener('submit', async (e) => {
        e.preventDefault();
        const id = document.getElementById('editId').value; 
        
        try {
            await fetch(`${API_BASE}/todo/update`, {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    id: id,
                    text: document.getElementById('editText').value,
                    date: document.getElementById('editDate').value,
                    time: document.getElementById('editTime').value
                })
            });
            closeModal();
            loadTasks();
        } catch (err) {
            console.error(err);
            alert("Failed to update task");
        }
    });
});

// NEW: Filter Logic
function setFilter(filterType) {
    currentFilter = filterType;
    
    // Update UI buttons
    document.querySelectorAll('.filter-btn').forEach(btn => btn.classList.remove('active'));
    event.target.classList.add('active');
    
    loadTasks(); // Reload to apply filter
}

async function loadTasks() {
    try {
        const res = await fetch(`${API_BASE}/todos`);
        if (!res.ok) throw new Error("Network response was not ok");
        let tasks = await res.json();
        
        // Apply Filter
        if (currentFilter === 'active') {
            tasks = tasks.filter(t => !t.done);
        } else if (currentFilter === 'completed') {
            tasks = tasks.filter(t => t.done);
        }
        
        renderTasks(tasks);
    } catch (error) {
        console.error("Fetch error:", error);
        document.getElementById('statsDisplay').innerText = "Connection Error";
    }
}

function renderTasks(tasks) {
    const listEl = document.getElementById('taskList');
    const emptyEl = document.getElementById('emptyState');
    const statsEl = document.getElementById('statsDisplay');

    listEl.innerHTML = ''; 
    
    // Calculate Stats based on ALL data (optional, or just filtered data)
    // Here we calculate based on what is currently shown
    let activeCount = 0;
    let totalCount = tasks.length;

    if (totalCount === 0) {
        emptyEl.style.display = 'block';
        // Dynamic Empty Message
        if(currentFilter === 'active') emptyEl.innerText = "No active tasks. Great job!";
        else if(currentFilter === 'completed') emptyEl.innerText = "No completed tasks yet.";
        else emptyEl.innerText = "Your schedule is clear.";
        
        statsEl.innerHTML = `<span class="active">0</span> Active / <span class="total">0</span> Total`;
        return;
    } else {
        emptyEl.style.display = 'none';
    }

    tasks.forEach(task => {
        if(!task.done) activeCount++;

        const li = document.createElement('li');
        li.className = `todo-row ${task.done ? 'is-done' : ''} ${task.is_late ? 'is-late' : ''}`;
        
        li.innerHTML = `
            <form action="#" onsubmit="event.preventDefault()">
                <input type="checkbox" class="custom-check" ${task.done ? 'checked' : ''} onchange="toggleTask('${task.id}')">
            </form>
            
            <div class="info-column">
                <span class="item-title">${escapeHtml(task.text)}</span>
                <div class="item-meta">
                    <span><svg class="meta-icon" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect><line x1="16" y1="2" x2="16" y2="6"></line><line x1="8" y1="2" x2="8" y2="6"></line><line x1="3" y1="10" x2="21" y2="10"></line></svg> ${task.date}</span>
                    <span><svg class="meta-icon" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"></circle><polyline points="12 6 12 12 16 14"></polyline></svg> ${task.time}</span>
                    ${task.is_late ? '<span class="badge-late">LATE</span>' : ''}
                </div>
            </div>

            <div class="action-column">
                <button class="icon-btn edit-icon" onclick="openEditModal('${task.id}', '${escapeJs(task.text)}', '${task.date}', '${task.time}')" title="Edit Task">
                    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path></svg>
                </button>
                
                <button class="icon-btn del-icon" onclick="deleteTask('${task.id}')" title="Delete Task">
                    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"></polyline><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path></svg>
                </button>
            </div>
        `;
        listEl.appendChild(li);
    });

    statsEl.innerHTML = `<span class="active">${activeCount}</span> Active / <span class="total">${totalCount}</span> Total`;
}

// Actions
async function toggleTask(id) {
    try {
        // 1. Send request to backend
        const response = await fetch(`${API_BASE}/todo/toggle`, {
            method: 'PATCH',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ id: id })
        });

        // 2. Check if response was successful (Status 200-299)
        if (!response.ok) {
            throw new Error(`Server responded with status ${response.status}`);
        }

        // 3. ONLY reload tasks if success confirmed
        console.log("Toggle successful, reloading...");
        await loadTasks(); 
        
    } catch (err) {
        console.error("Toggle failed:", err);
        alert("Failed to update task status. Please check terminal logs.");
        
        // Optional: Revert visual state if it fails? 
        // For now, just let the next manual refresh fix it.
    }
}

// Modal Logic
function openEditModal(id, text, date, time) {
    document.getElementById('editId').value = id;
    document.getElementById('editText').value = text;
    document.getElementById('editDate').value = date;
    document.getElementById('editTime').value = time;
    document.getElementById('editModal').style.display = 'flex';
}

function closeModal() {
    document.getElementById('editModal').style.display = 'none';
}

// Helpers
function escapeHtml(text) {
    const map = {'&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#039;'};
    return text.replace(/[&<>"']/g, m => map[m]);
}
function escapeJs(str) {
    return str.replace(/\\/g, '\\\\').replace(/'/g, "\\'").replace(/"/g, '\\"');
}