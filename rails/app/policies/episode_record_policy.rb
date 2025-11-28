# typed: false
# frozen_string_literal: true

class EpisodeRecordPolicy < ApplicationPolicy
  def update?
    user == record.user
  end

  def destroy?
    user == record.user
  end
end
