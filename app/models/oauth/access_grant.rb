# frozen_string_literal: true

# == Schema Information
#
# Table name: oauth_access_grants
#
#  id                :bigint           not null, primary key
#  expires_in        :integer          not null
#  redirect_uri      :text             not null
#  revoked_at        :datetime
#  scopes            :string
#  token             :string           not null
#  created_at        :datetime         not null
#  application_id    :bigint           not null
#  resource_owner_id :bigint           not null
#
# Indexes
#
#  index_oauth_access_grants_on_token  (token) UNIQUE
#
# Foreign Keys
#
#  fk_rails_...  (application_id => oauth_applications.id)
#  fk_rails_...  (resource_owner_id => users.id)
#
class Oauth::AccessGrant < ApplicationRecord
  include ::Doorkeeper::Orm::ActiveRecord::Mixins::AccessGrant

  include BatchDestroyable
end
