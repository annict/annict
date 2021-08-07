Module.class_eval do
  ##
  # 独自のエレメントモジュールなどをincludeする処理のショートカット
  #
  # 例:
  # module MyElement
  #   rspec type: :request
  #
  #   matcher :header_menu do |alt|
  #     ...
  #   end
  # end
  #
  def rspec(options = {})
    RSpec.configure do |config|
      config.include(self, options)
    end
  end
end
