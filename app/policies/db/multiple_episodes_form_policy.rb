class DB::MultipleEpisodesFormPolicy < ApplicationPolicy
  def create?
    user.role.editor? || user.role.admin?
  end
end
