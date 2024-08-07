# typed: true

# DO NOT EDIT MANUALLY
# This is an autogenerated file for dynamic methods in `Api::Internal::RecordsController`.
# Please instead update this file by running `bin/tapioca dsl Api::Internal::RecordsController`.


class Api::Internal::RecordsController
  sig { returns(HelperProxy) }
  def helpers; end

  module HelperMethods
    include ::Ransack::Helpers::FormHelper
    include ::ActionController::Base::HelperMethods
    include ::ApplicationHelper
    include ::ComponentValueFetcherHelper
    include ::Db::ApplicationHelper
    include ::EpisodesHelper
    include ::GaHelper
    include ::HeadHelper
    include ::IcalendarHelper
    include ::IconHelper
    include ::ImageHelper
    include ::LocalHelper
    include ::MarkdownHelper
    include ::RecordsHelper
    include ::StaffsHelper
    include ::TimeZoneHelper
    include ::UrlHelper
    include ::VodHelper
    include ::WorksHelper
    include ::PreviewHelper
    include ::Doorkeeper::DashboardHelper
    include ::DeviseHelper
    include ::ApplicationV6Controller::HelperMethods
    include ::Pundit::Helper

    sig { params(record: T.untyped).returns(T.untyped) }
    def policy(record); end

    sig { params(scope: T.untyped).returns(T.untyped) }
    def pundit_policy_scope(scope); end

    sig { returns(T.untyped) }
    def pundit_user; end
  end

  class HelperProxy < ::ActionView::Base
    include HelperMethods
  end
end
