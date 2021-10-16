# frozen_string_literal: true

module Forms
  class SupporterForm < Forms::ApplicationForm
    include Supportable

    attr_writer :gumroad_subscriber_id

    def subscriber
      @subscriber ||= gumroad_client.fetch_subscriber_by_subscriber_id(@gumroad_subscriber_id)
    end
  end
end
