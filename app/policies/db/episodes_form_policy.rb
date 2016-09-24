# frozen_string_literal: true

module DB
  class EpisodesFormPolicy < ApplicationPolicy
    def edit?
      user.committer?
    end

    def update?
      user.committer?
    end
  end
end
