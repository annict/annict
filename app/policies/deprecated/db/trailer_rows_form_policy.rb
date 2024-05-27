# typed: false
# frozen_string_literal: true

module Deprecated::Db
  class TrailerRowsFormPolicy < ApplicationPolicy
    def create?
      user.present? && user.committer?
    end

    def update?
      user.present? && user.committer?
    end
  end
end
