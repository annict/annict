# typed: false
# frozen_string_literal: true

module ResourcePublishable
  extend ActiveSupport::Concern

  def create
    authorize(create_resource, :publish?)

    create_resource.publish

    redirect_back(
      fallback_location: after_created_path,
      notice: t("messages._common.published")
    )
  end

  def destroy
    authorize(destroy_resource, :unpublish?)

    destroy_resource.unpublish

    redirect_back(
      fallback_location: after_destroyed_path,
      notice: t("messages._common.unpublished")
    )
  end

  private

  def create_resource
    raise NotImplementedError
  end

  def destroy_resource
    raise NotImplementedError
  end

  def after_created_path
    raise NotImplementedError
  end

  def after_destroyed_path
    raise NotImplementedError
  end
end
