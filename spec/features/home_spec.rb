require 'spec_helper'

describe 'トップページ' do

  context 'ログインしていないとき' do
    let(:work) { create(:work, :with_item) }
    let(:cover_image) { create(:cover_image, work: work) }

    before do
      visit '/'
    end

    it 'アクセスするとページが表示される' do
      expect(page).to have_content('見たアニメを記録して、共有しよう')
    end

    it 'Twitterでログインをクリックするとユーザ登録ページが表示される' do
      find('.welcome').click_link('Twitterアカウントで始める')

      expect(current_path).to eq '/users/sign_up'
    end

    it 'Facebookでログインをクリックするとユーザ登録ページが表示される' do
      find('.welcome').click_link('Facebookアカウントでログイン')

      expect(current_path).to eq '/users/sign_up'
    end
  end

  context 'ログインしているとき' do
    let(:user) { create(:registered_user) }

    before do
      login_as(user, scope: :user)
    end

    context 'アクティビティが存在しないとき' do
      before do
        visit '/'
      end

      it 'アクティビティが存在しない旨を表示すること', js: true do
        expect(page).to have_content('アクティビティはありませんでした')
      end
    end

    context '自分がチェックインしているとき' do
      let!(:checkin) { create(:checkin, user: user, comment: 'おもしろかったよ') }

      before do
        visit '/'
      end

      it 'アクティビティに自分のチェックイン情報が表示されること', js: true do
        expect(page).to have_content('おもしろかったよ')
      end
    end

    context 'フォローしている人がチェックインしているとき' do
      let!(:following_user) { create(:registered_user) }
      let!(:checkin) { create(:checkin, user: following_user, comment: 'たのしかったよ') }

      before do
        user.follow(following_user)

        visit '/'
      end

      it 'アクティビティにフォローしている人のチェックイン情報が表示されること', js: true do
        expect(page).to have_content('たのしかったよ')
      end
    end
  end
end
