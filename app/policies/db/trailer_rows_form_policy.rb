# frozen_string_literal: true

module Db
  class TrailerRowsFormPolicy < ApplicationPolicy
    def create?
      user.committer?
    end

    def update?
      user.committer?
    end
  end
end
