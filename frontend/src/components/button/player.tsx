import { Player as LottiePlayer } from "@lottiefiles/react-lottie-player";
export default function Player(props: any) {
  // @ts-expect-error
  return <LottiePlayer {...props} />;
}
