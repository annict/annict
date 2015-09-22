Capybara.default_max_wait_time = 8 # Seconds to wait before timeout error. Default is 2

# Register slightly larger than default window size...
Capybara.register_driver :poltergeist do |app|
  Capybara::Poltergeist::Driver.new(app, {
    debug: false,              # change this to true to troubleshoot
    window_size: [1300, 1000], # this can affect dynamic layout
    timeout: 60,
    js_errors: false
  })
end

Capybara.javascript_driver = :poltergeist

##
# スクショ画像を作成するためのメソッド。`tmp/render` 以下に画像が保存されます。
# 注意: このメソッドを使用するときは、`js` オプションを `context` メソッドなどに指定してください。
#
# 例:
# describe '#index' do
#   context 'うまくいってるとき', js: true do
#     render_page('すくしょ')
#   end
# end
#
def render_page(name)
  png_name = name.strip.gsub(/\W+/, '-')
  path = File.join('tmp/render', "#{png_name}.png")
  save_screenshot(path)
end

# HTML形式でページを保存する `save_and_open_page` メソッドのショートカット
def save!
  save_and_open_page
end
