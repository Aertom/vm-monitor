import { useMemo, useState } from 'react';
import type { Group } from '../types/vm';
import { FamilyBadge } from './FamilyBadge';
import { useCheckout } from '../hooks/useCheckout';

interface GroupTableProps {
  groups: Group[];
  isLoading: boolean;
}

type SortKey = 'id' | 'status';

const FAMILIES = ['sm', 'cm', 'ws', 'oa'] as const;

export function GroupTable({ groups, isLoading }: GroupTableProps) {
  const [familyFilter, setFamilyFilter] = useState<string>('all');
  const [statusFilter, setStatusFilter] = useState<'all' | 'in-use' | 'free'>('all');
  const [sortKey, setSortKey] = useState<SortKey>('id');
  const [nameDrafts, setNameDrafts] = useState<Record<string, string>>({});

  const { checkout, checkin, isPending } = useCheckout();

  const filteredGroups = useMemo(() => {
    let result = groups;

    if (familyFilter !== 'all') {
      result = result.filter((g) => g.vms[familyFilter] !== undefined);
    }

    if (statusFilter === 'in-use') {
      result = result.filter((g) => !!g.in_use_by);
    } else if (statusFilter === 'free') {
      result = result.filter((g) => !g.in_use_by);
    }

    return [...result].sort((a, b) => {
      if (sortKey === 'status') {
        const aInUse = a.in_use_by ? 1 : 0;
        const bInUse = b.in_use_by ? 1 : 0;
        if (aInUse !== bInUse) return bInUse - aInUse;
      }
      return a.id.localeCompare(b.id);
    });
  }, [groups, familyFilter, statusFilter, sortKey]);

  if (isLoading) {
    return <div className="state-panel loading">Chargement des groupes de VMs…</div>;
  }

  if (groups.length === 0) {
    return (
      <div className="state-panel empty">
        Aucun groupe détecté pour le moment. La découverte automatique ou la
        collecte SSH n'a pas encore trouvé de VM.
      </div>
    );
  }

  return (
    <div className="group-table-wrapper">
      <div className="toolbar">
        <label>
          Famille :
          <select value={familyFilter} onChange={(e) => setFamilyFilter(e.target.value)}>
            <option value="all">Toutes</option>
            {FAMILIES.map((f) => (
              <option key={f} value={f}>
                {f.toUpperCase()}
              </option>
            ))}
          </select>
        </label>

        <label>
          Statut :
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value as typeof statusFilter)}
          >
            <option value="all">Tous</option>
            <option value="in-use">En cours d'utilisation</option>
            <option value="free">Libres</option>
          </select>
        </label>

        <label>
          Trier par :
          <select value={sortKey} onChange={(e) => setSortKey(e.target.value as SortKey)}>
            <option value="id">Nom du groupe</option>
            <option value="status">Statut (en cours d'abord)</option>
          </select>
        </label>

        <span className="result-count">
          {filteredGroups.length} groupe{filteredGroups.length > 1 ? 's' : ''}
        </span>
      </div>

      {filteredGroups.length === 0 ? (
        <div className="state-panel empty">Aucun groupe ne correspond aux filtres.</div>
      ) : (
        <table className="group-table">
          <thead>
            <tr>
              <th>Groupe</th>
              {FAMILIES.map((f) => (
                <th key={f}>
                  <FamilyBadge family={f} />
                </th>
              ))}
              <th>Statut</th>
              <th>Action</th>
            </tr>
          </thead>
          <tbody>
            {filteredGroups.map((g) => (
              <tr key={g.id} className={g.in_use_by ? 'row-in-use' : ''}>
                <td className="group-id">{g.id}</td>
                {FAMILIES.map((f) => {
                  const vm = g.vms[f];
                  return (
                    <td key={f} className="vm-cell">
                      {vm ? (
                        <div className="vm-info">
                          <div className="vm-hostname">{vm.hostname}</div>
                          <div className="vm-ip">{vm.ip}</div>
                          <div className="vm-hypervisor">{vm.hypervisor}</div>
                          {vm.app_versions &&
                            Object.entries(vm.app_versions).map(([app, version]) => (
                              <div key={app} className="vm-app-version">
                                {app} <span className="version">{version}</span>
                              </div>
                            ))}
                          {vm.last_error && (
                            <div className="vm-error" title={vm.last_error}>
                              ⚠ erreur de collecte
                            </div>
                          )}
                        </div>
                      ) : (
                        <span className="vm-missing">—</span>
                      )}
                    </td>
                  );
                })}
                <td>
                  {g.in_use_by ? (
                    <span className="status-badge in-use" title={g.checked_out_at ?? ''}>
                      🟠 utilisé par <strong>{g.in_use_by}</strong>
                    </span>
                  ) : (
                    <span className="status-badge free">🟢 libre</span>
                  )}
                </td>
                <td>
                  {g.in_use_by ? (
                    <button
                      className="btn btn-checkin"
                      disabled={isPending}
                      onClick={() => checkin(g.id)}
                    >
                      Libérer
                    </button>
                  ) : (
                    <div className="checkout-form">
                      <input
                        type="text"
                        placeholder="Votre nom"
                        value={nameDrafts[g.id] ?? ''}
                        onChange={(e) =>
                          setNameDrafts((prev) => ({ ...prev, [g.id]: e.target.value }))
                        }
                      />
                      <button
                        className="btn btn-checkout"
                        disabled={isPending || !(nameDrafts[g.id] ?? '').trim()}
                        onClick={() => checkout(g.id, (nameDrafts[g.id] ?? '').trim())}
                      >
                        Utiliser
                      </button>
                    </div>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}
