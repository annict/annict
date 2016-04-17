# frozen_string_literal: true

describe "キーワード検索ページ" do
  let!(:work) { create(:work, :with_item) }

  before do
    visit "/search"

    within(".form-group") do
      fill_in "q", with: work.title
    end
    click_button "検索"
  end

  it "検索結果に該当の作品が表示されること" do
    expect(find(".app__main .works")).to have_content(work.title)
  end
end
