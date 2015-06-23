class EpisodePolicy < ApplicationPolicy
  def destroy?
    user.role.admin?
  end
end
