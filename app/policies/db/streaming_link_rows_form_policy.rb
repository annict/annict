# frozen_string_literal: true

module DB
  class StreamingLinkRowsFormPolicy < ApplicationPolicy
    def create?
      user.committer?
    end

    def update?
      user.committer?
    end
  end
end
