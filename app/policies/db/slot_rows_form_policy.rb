# frozen_string_literal: true

module Db
  class SlotRowsFormPolicy < ApplicationPolicy
    def create?
      user.committer?
    end

    def update?
      user.committer?
    end
  end
end
