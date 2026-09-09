# vm-monitor

Application de monitoring d'un parc de VM déployées sur des hyperviseurs hétérogènes (ESXi, AHV, KVM), toutes dans le même sous-réseau, accessibles en SSH par clé.

## Fonctionnalités

- Collecte des infos VM (IP, hostname, hyperviseur, versions d'applis installées) via SSH
- 4 familles de VM identifiées par mot-clé dans le hostname : `sm`, `cm`, `ws`, `oa`
  - `sm` : versions cherchées sous `/opt/appli_x.y.z`
  - `cm`, `ws`, `oa` : versions cherchées sous `/appli/appli_x.y.z`
- Détection des groupes de VM liées via le contenu de `/etc/hosts` sur chaque VM
- Tableaux de bord par groupe (IP, hostname, hyperviseur, versions)
- Checkout "nom libre" par un utilisateur pour signaler qu'il utilise un groupe (stockage en mémoire pour l'instant)

## Statut actuel

- [x] Inventaire des VMs en statique (fichier YAML)
- [x] Collecteur SSH (versions d'applis, /etc/hosts)
- [x] Scheduler de collecte périodique
- [x] Store en mémoire (VMs, groupes, checkout)
- [x] API HTTP (groupes, checkout/checkin)
- [x] Squelette frontend TypeScript
- [ ] Découverte automatique via l'API ESXi/vCenter (govmomi) — à venir
- [ ] Découverte automatique via l'API AHV/Prism — à venir
- [ ] Découverte automatique via libvirt (KVM) — à venir

## Structure

```
backend/    -> API Go + collecteur SSH + scheduler
frontend/   -> Application TypeScript (React + Vite)
```

## Démarrage rapide (backend)

```bash
cd backend
cp config/inventory.example.yaml config/inventory.yaml
# éditer inventory.yaml avec vos VMs et le chemin de la clé SSH
go run ./cmd/server
```

Variables d'environnement optionnelles :
- `VMMONITOR_CONFIG` (défaut: `config/inventory.yaml`)
- `VMMONITOR_ADDR` (défaut: `:8080`)
