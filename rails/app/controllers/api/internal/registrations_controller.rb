# typed: false
# frozen_string_literal: true

module Api::Internal
  class RegistrationsController < Api::Internal::ApplicationController
    def create
      @confirmation = EmailConfirmation.find_by!(event: :sign_up, token: registration_form_params[:token])

      @form = Forms::RegistrationForm.new(registration_form_params.merge(email: @confirmation.email))

      if @form.invalid?
        return render json: @form.errors.full_messages, status: :unprocessable_entity
      end

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
      render(
        json: {redirect_path: @confirmation.back.presence || root_path},
        status: 201
      )
    end

    private

    def registration_form_params
      params.require(:forms_registration_form).permit(:email, :terms_and_privacy_policy_agreement, :token, :username)
    end
  end
end
