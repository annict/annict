# frozen_string_literal: true

module V4
  class RegistrationsController < V4::ApplicationController
    layout "simple"

    def new
      redirect_if_signed_in

      token = params[:token]

      unless token
        return redirect_to root_path
      end

      confirmation = EmailConfirmation.find_by(event: :sign_up, token: token)

      if !confirmation || confirmation.expired?
        @expired = true
        return
      end

      confirmation.touch(:expires_at)

      @form = RegistrationForm.new
      @form.email = confirmation.email
      @form.token = confirmation.token
    end

    def create
      redirect_if_signed_in

      token = registration_form_attributes[:token]
      @confirmation = EmailConfirmation.find_by(event: :sign_up, token: token)

      unless @confirmation
        return redirect_to root_path
      end

      @form = RegistrationForm.new(registration_form_attributes.merge(email: @confirmation.email))

      return render(:new) unless @form.valid?

      user = User.new(
        username: @form.username,
        email: @form.email
      ).build_relations
      user.time_zone = cookies["ann_time_zone"].presence || "Asia/Tokyo"
      user.locale = locale
      user.confirmed_at = Time.zone.now
      user.setting.privacy_policy_agreed = true

      ActiveRecord::Base.transaction do
        user.save!
        @confirmation.destroy

        sign_in user
      end

      flash[:notice] = t("messages.registrations.create.welcome")
      redirect_to(@confirmation.back.presence || root_path)
    end

    private

    def registration_form_attributes
      @registration_form_attributes ||= params.to_unsafe_h["registration_form"]
    end
  end
end
