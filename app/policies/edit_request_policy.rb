class EditRequestPolicy < ApplicationPolicy
  def publish?
    user.present? && user.committer?
  end

  def close?
    user.present? && (user.committer? || user == record.user)
  end
end
