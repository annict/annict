# typed: false
# frozen_string_literal: true

class MemberConstraint
  def matches?(_request)
    # !User.without_deleted.find_by_session(request.session).nil?
    false
  end
end

class GuestConstraint
  def matches?(request)
    !MemberConstraint.new.matches?(request)
  end
end

root "home#show", constraints: MemberConstraint.new
root "welcome#show", constraints: GuestConstraint.new, as: nil # Set :as option to avoid two routes with the same name
