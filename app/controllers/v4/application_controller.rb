# frozen_string_literal: true

module V4
  class ApplicationController < ActionController::Base
    include V4::RavenContext
    include V4::Loggable
    include V4::Localizable
    include V4::PageCategorizable

    layout "v4"

    helper_method :client_uuid, :local_url_with_path, :locale_en?, :locale_ja?, :local_url, :page_category

    before_action :set_raven_context
    around_action :set_locale

    private

    # Override `Devise::Controllers::Helpers#signed_in_root_path`
    def signed_in_root_path(_resource_or_scope)
      root_path
    end

    # Override `Devise::Controllers::Helpers#after_sign_out_path_for`
    def after_sign_out_path_for(_resource_or_scope)
      root_path
    end

    def graphql_client
      @graphql_client ||= Annict::Graphql::InternalClient.new(
        # We don't fetch user related data from GraphQL to cache GraphQL data.
        viewer: nil
      )
    end
  end
end
