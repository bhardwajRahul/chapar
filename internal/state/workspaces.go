package state

import (
	"github.com/chapar-rest/chapar/internal/domain"
	"github.com/chapar-rest/chapar/internal/repository"
	"github.com/chapar-rest/chapar/internal/safemap"
)

type (
	WorkspaceChangeListener       func(workspace *domain.Workspace, source Source, action Action)
	ActiveWorkspaceChangeListener func(*domain.Workspace)
)

type Workspaces struct {
	workspaceChangeListeners       []WorkspaceChangeListener
	activeWorkspaceChangeListeners []ActiveWorkspaceChangeListener
	workspaces                     *safemap.Map[*domain.Workspace]

	activeWorkspace *domain.Workspace
	repository      repository.RepositoryV2
}

func NewWorkspaces(repository repository.RepositoryV2) (*Workspaces, error) {
	// NOTE(Isaac799) if no workspaces exist we create one to prevent nil deref
	// on startup when header getting selected workspace ddl option
	workspaces, err := repository.LoadWorkspaces()
	if err != nil {
		return nil, err
	}
	if len(workspaces) == 0 {
		workspace := domain.NewWorkspace(domain.DefaultWorkspaceName)
		err := repository.CreateWorkspace(workspace)
		if err != nil {
			return nil, err
		}
	}

	workspace := Workspaces{
		repository: repository,
		workspaces: safemap.New[*domain.Workspace](),
	}

	return &workspace, nil
}

func (m *Workspaces) AddWorkspaceChangeListener(listener WorkspaceChangeListener) {
	m.workspaceChangeListeners = append(m.workspaceChangeListeners, listener)
}

func (m *Workspaces) AddActiveWorkspaceChangeListener(listener ActiveWorkspaceChangeListener) {
	m.activeWorkspaceChangeListeners = append(m.activeWorkspaceChangeListeners, listener)
}

func (m *Workspaces) notifyWorkspaceChange(workspace *domain.Workspace, source Source, action Action) {
	for _, listener := range m.workspaceChangeListeners {
		listener(workspace, source, action)
	}
}

func (m *Workspaces) notifyActiveWorkspaceChange(workspace *domain.Workspace) {
	for _, listener := range m.activeWorkspaceChangeListeners {
		listener(workspace)
	}
}

func (m *Workspaces) AddWorkspace(workspace *domain.Workspace, source Source) {
	m.workspaces.Set(workspace.MetaData.ID, workspace)
	m.notifyWorkspaceChange(workspace, source, ActionAdd)
}

func (m *Workspaces) RemoveWorkspace(workspace *domain.Workspace, source Source, stateOnly bool) error {
	if _, ok := m.workspaces.Get(workspace.MetaData.ID); !ok {
		return ErrNotFound
	}

	if !stateOnly {
		if err := m.repository.DeleteWorkspace(workspace); err != nil {
			return err
		}
	}

	m.workspaces.Delete(workspace.MetaData.ID)
	m.notifyWorkspaceChange(workspace, source, ActionDelete)
	return nil
}

func (m *Workspaces) GetWorkspace(id string) *domain.Workspace {
	ws, _ := m.workspaces.Get(id)
	return ws
}

func (m *Workspaces) UpdateWorkspace(workspace *domain.Workspace, source Source, stateOnly bool) error {
	if _, ok := m.workspaces.Get(workspace.MetaData.ID); !ok {
		return ErrNotFound
	}

	if !stateOnly {
		if err := m.repository.UpdateWorkspace(workspace); err != nil {
			return err
		}
	}

	m.workspaces.Set(workspace.MetaData.ID, workspace)
	m.notifyWorkspaceChange(workspace, source, ActionUpdate)

	return nil
}

func (m *Workspaces) SetActiveWorkspace(workspace *domain.Workspace) {
	if _, ok := m.workspaces.Get(workspace.MetaData.ID); !ok {
		return
	}

	m.activeWorkspace = workspace
	m.notifyActiveWorkspaceChange(workspace)
}

func (m *Workspaces) GetActiveWorkspace() *domain.Workspace {
	return m.activeWorkspace
}

func (m *Workspaces) ClearActiveWorkspace() {
	m.activeWorkspace = nil
	m.notifyActiveWorkspaceChange(nil)
}

func (m *Workspaces) GetWorkspaces() []*domain.Workspace {
	return m.workspaces.Values()
}

func (m *Workspaces) LoadWorkspaces() ([]*domain.Workspace, error) {
	ws, err := m.repository.LoadWorkspaces()
	if err != nil {
		return nil, err
	}

	for _, w := range ws {
		m.workspaces.Set(w.MetaData.ID, w)
	}

	return ws, nil
}
