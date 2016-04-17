# frozen_string_literal: true

describe "今期作品一覧ページ" do
  let!(:work) { create(:work, :with_item, :with_current_season) }

  before do
    visit "/works"
  end

  it "今期の作品が表示されること" do
    expect(page).to have_content(work.title)
  end
end

describe "クールごとの作品一覧ページ" do
  let!(:season1) { create(:season) }
  let!(:season2) { create(:season) }
  let!(:work1) { create(:work, :with_item, season: season1) }
  let!(:work2) { create(:work, :with_item, season: season2) }

  before do
    visit "/works/#{season1.slug}"
  end

  it "クールごとの作品が表示されること" do
    expect(page).to have_content(work1.title)
    expect(page).to_not have_content(work2.title)
  end
end

describe "人気の作品一覧ページ" do
  let!(:work) { create(:work, :with_item) }

  before do
    visit "/works/popular"
  end

  it "人気の作品が表示されること" do
    expect(page).to have_content(work.title)
  end
end

describe "作品詳細ページ" do
  let!(:work) { create(:work, :with_item) }

  context "ログインしているとき" do
    let(:user) { create(:registered_user) }

    before do
      login_as(user, scope: :user)
    end

    context "作品ページにアクセスしたとき" do
      before do
        visit "/works/#{work.id}"
      end

      it "ページが表示される" do
        expect(page).to have_content(work.title)
      end
    end
  end
end
