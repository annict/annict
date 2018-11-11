# frozen_string_literal: true

module AssetsHelper
  class BundleNotFound < StandardError; end

  def asset_bundle_path(entry, **options)
    raise BundleNotFound, "Could not find bundle with name #{entry}" unless manifest.key?(entry)

    asset_path(manifest.fetch(entry), **options)
  end

  def asset_bundle_url(entry, options = {})
    raise BundleNotFound, "Could not find bundle with name #{entry}" unless manifest.key?(entry)

    options = options.merge(host: ENV.fetch("ANNICT_ASSET_URL"))
    asset_url(manifest.fetch(entry), options)
  end

  def javascript_bundle_tag(entry, **options)
    javascript_include_tag(asset_bundle_path(entry), **options)
  end

  def stylesheet_bundle_tag(entry, **options)
    stylesheet_link_tag(asset_bundle_path(entry), **options)
  end

  def image_bundle_tag(entry, **options)
    image_tag(asset_bundle_path(entry), **options)
  end

  private

  MANIFEST_PATH = "public/packs/manifest.json"

  def manifest
    @manifest ||= JSON.parse(File.read(MANIFEST_PATH))
  end
end
