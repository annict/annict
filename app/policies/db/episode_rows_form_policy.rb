# frozen_string_literal: true

module DB
  class EpisodeRowsFormPolicy < ApplicationPolicy
    def create?
      user.committer?
    end

    def update?
      user.committer?
    end
  end
end
