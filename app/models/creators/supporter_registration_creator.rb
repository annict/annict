# frozen_string_literal: true

module Creators
  class SupporterRegistrationCreator
    def initialize(user:, form:)
      @user = user
      @form = form
    end

    def call
      ActiveRecord::Base.transaction do
        provider = @user.providers.where(name: @form.provider_name).first_or_initialize
        provider.uid = @form.provider_uid
        provider.token = @form.provider_token
        provider.save!

        gs = GumroadSubscriber.where(gumroad_id: @form.gumroad_subscriber_id).first_or_initialize
        gs.gumroad_product_id = @form.gumroad_product_id
        gs.gumroad_product_name = @form.gumroad_product_name
        gs.gumroad_user_id = @form.gumroad_user_id
        gs.gumroad_user_email = @form.gumroad_user_email
        gs.gumroad_purchase_ids = @form.gumroad_purchase_ids
        gs.gumroad_created_at = @form.gumroad_created_at
        gs.gumroad_cancelled_at = @form.gumroad_cancelled_at
        gs.gumroad_user_requested_cancellation_at = @form.gumroad_user_requested_cancellation_at
        gs.gumroad_charge_occurrence_count = @form.gumroad_charge_occurrence_count
        gs.gumroad_ended_at = @form.gumroad_ended_at
        gs.save!

        @user.update!(gumroad_subscriber: gs)
      end
    end
  end
end
