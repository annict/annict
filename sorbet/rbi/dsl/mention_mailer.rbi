# typed: true

# DO NOT EDIT MANUALLY
# This is an autogenerated file for dynamic methods in `MentionMailer`.
# Please instead update this file by running `bin/tapioca dsl MentionMailer`.


class MentionMailer
  class << self
    sig do
      params(
        username: T.untyped,
        resource_id: T.untyped,
        resource_type: T.untyped,
        column: T.untyped
      ).returns(::ActionMailer::MessageDelivery)
    end
    def notify(username, resource_id, resource_type, column); end
  end
end
