# frozen_string_literal: true
# == Schema Information
#
# Table name: organizations
#
#  id               :integer          not null, primary key
#  name             :string           not null
#  url              :string
#  wikipedia_url    :string
#  twitter_username :string
#  aasm_state       :string           default("published"), not null
#  created_at       :datetime         not null
#  updated_at       :datetime         not null
#
# Indexes
#
#  index_organizations_on_aasm_state  (aasm_state)
#  index_organizations_on_name        (name) UNIQUE
#

class OrganizationsController < ApplicationController
  def show(id)
    @organization = Organization.published.find(id)
    @staffs_with_year = @organization.
      staffs.
      published.
      # includes(work: [:season, :item]).
      group_by { |s| s.work.season&.year.presence || 0 }
    @staff_years = @staffs_with_year.keys.sort.reverse

    render layout: "v1/application"
  end
end
