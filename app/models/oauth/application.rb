# frozen_string_literal: true

# == Schema Information
#
# Table name: oauth_applications
#
#  id                :bigint           not null, primary key
#  aasm_state        :string           default("published"), not null
#  confidential      :boolean          default(TRUE), not null
#  deleted_at        :datetime
#  hide_social_login :boolean          default(FALSE), not null
#  name              :string           not null
#  owner_type        :string
#  redirect_uri      :text             not null
#  scopes            :string           default(""), not null
#  secret            :string           not null
#  uid               :string           not null
#  created_at        :datetime
#  updated_at        :datetime
#  owner_id          :bigint
#
# Indexes
#
#  index_oauth_applications_on_deleted_at               (deleted_at)
#  index_oauth_applications_on_owner_id_and_owner_type  (owner_id,owner_type)
#  index_oauth_applications_on_uid                      (uid) UNIQUE
#
class Oauth::Application < ApplicationRecord
  include ::Doorkeeper::Orm::ActiveRecord::Mixins::Application

  include BatchDestroyable

  scope :available, -> { where(deleted_at: nil).where.not(owner: nil) }
  scope :unavailable, -> {
    unscoped.where.not(deleted_at: nil).or(where(owner: nil))
  }
  scope :authorized, -> { where(oauth_access_tokens: {revoked_at: nil}) }
end
