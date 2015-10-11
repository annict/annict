class EpisodePolicy < ApplicationPolicy
  def update?
    user.role.editor? || user.role.admin?
  end

  def hide?
    user.role.admin?
  end

  def destroy?
    user.role.admin?
  end
end
