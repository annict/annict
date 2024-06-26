# typed: true

# DO NOT EDIT MANUALLY
# This is an autogenerated file for dynamic methods in `EmailConfirmationMailer`.
# Please instead update this file by running `bin/tapioca dsl EmailConfirmationMailer`.


class EmailConfirmationMailer
  class << self
    sig { params(locale: T.untyped).returns(::ActionMailer::MessageDelivery) }
    def local_domain(locale: T.unsafe(nil)); end

    sig { params(locale: T.untyped).returns(::ActionMailer::MessageDelivery) }
    def local_url(locale: T.unsafe(nil)); end

    sig { params(email_confirmation_id: T.untyped, locale: T.untyped).returns(::ActionMailer::MessageDelivery) }
    def sign_in_confirmation(email_confirmation_id, locale); end

    sig { params(email_confirmation_id: T.untyped, locale: T.untyped).returns(::ActionMailer::MessageDelivery) }
    def sign_up_confirmation(email_confirmation_id, locale); end

    sig { params(email_confirmation_id: T.untyped, locale: T.untyped).returns(::ActionMailer::MessageDelivery) }
    def update_email_confirmation(email_confirmation_id, locale); end
  end
end
