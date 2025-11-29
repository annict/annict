# typed: false
# frozen_string_literal: true

module Deprecated::Db
  class OrganizationRowsFormPolicy < ApplicationPolicy
    def create?
      user.present? && user.committer?
    end

    def update?
      user.present? && user.committer?
    end
  end
end
