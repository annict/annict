# typed: false
# frozen_string_literal: true

module Deprecated::Db
  class CastRowsFormPolicy < ApplicationPolicy
    def create?
      user.present? && user.committer?
    end

    def update?
      user.present? && user.committer?
    end
  end
end
