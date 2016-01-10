class DB::MultipleEpisodesFormPolicy < ApplicationPolicy
  def create?
    user.committer?
  end
end
