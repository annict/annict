# frozen_string_literal: true

module Db
  class SeriesWorkRowsFormPolicy < ApplicationPolicy
    def create?
      user.committer?
    end

    def update?
      user.committer?
    end
  end
end
