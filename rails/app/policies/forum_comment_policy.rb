# typed: false
# frozen_string_literal: true

class ForumCommentPolicy < ApplicationPolicy
  def update?
    user == record.user
  end

  def destroy?
    user == record.user
  end
end
