# oa Architecture: Zero-Copy Filesystem

The core philosophy of **oa** is surgical efficiency. Traditional remastering tools spend hours performing a physical copy (`cp -a`) of the host filesystem. **oa** eliminates this bottleneck by implementing a **Zero-Copy** strategy powered by the Linux Kernel's **OverlayFS**.

---

## 🏗️ The Three-Layer Logic

Instead of duplicating data, **oa** creates a "projected" filesystem by stacking distinct layers. This approach, perfected over years in the *penguins-eggs* project, is now implemented in native C for maximum performance.

### 1. LowerDir (The Host - Read-Only)
* **Source**: The root filesystem (`/`) of the running host.
* **State**: Mounted as **Strictly Read-Only**.
* **Impact**: Zero disk I/O, zero time consumed. It serves as the immutable "DNA" of the new ISO.

### 2. UpperDir (The Workspace - Read/Write)
* **Source**: A dedicated directory in the work path (e.g., `.overlay/upperdir`).
* **State**: Writable.
* **Role**: This layer captures the "delta." Any modification (adding users, deleting logs, changing configs) is stored here without ever touching the host.

### 3. MergedDir (The Liveroot)
* **Role**: The unified mount point.
* **Appearance**: To the engine and tools (`chroot`, `mksquashfs`), it looks like a standard, writable root filesystem.
* **Mechanism**: The Kernel merges the Host and the Workspace in real-time.

---

## 🛡️ Advanced Isolation & Protections

To ensure a "Zero-Footprint" operation and avoid system crashes, **oa** implements advanced kernel-level safeguards during the `prepare` phase:

1. **Mount Namespace Isolation**: Uses `MS_PRIVATE` propagation to ensure that mounts inside the `liveroot` do not "leak" back to the host system.
2. **Kernel API Projection**: Critical interfaces (`/dev`, `/proc`, `/sys`, `/run`) are bind-mounted from the host into the `merged` directory to allow system tools to function normally.
3. **Smart `/home` Handling**: The treatment of user data adapts to the operation mode:
    * In **Clone** or **Crypted** modes, the host's `/home` is bind-mounted securely in read-only mode.
    * In **Standard** mode, `/home` is created as a completely empty directory to safely host the newly crafted live user, leaving host data perfectly isolated.
4. **Anti-Inception Mask (tmpfs)**: When scanning or compressing the `liveroot`, there is a high risk of infinite recursive loops if the workspace is located inside the projected filesystem (e.g., inside `/home`). **oa** prevents this by dynamically mounting an empty `tmpfs` layer precisely over the workspace path *inside* the `liveroot`, completely hiding it from tools like `mksquashfs` or `nftw`.

---

## 🚀 Key Advantages

* **Instant Setup**: The `liveroot` is ready for customization in milliseconds, regardless of the host size (10GB or 1TB).
* **SSD Longevity**: Eliminates Gigabytes of unnecessary write cycles.
* **Physical Safety**: Since the Host (LowerDir) is Read-Only, the running system is physically shielded from accidental damage during the remastering process.
* **Atomic Cleanup**: Removing the `liveroot` simply requires unmounting the overlays (using `MNT_DETACH`) and safely dropping the temporary directories.

---
*Documented by Piero Proietti - 2026*