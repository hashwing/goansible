package transport

// func (s *Session) Start(cmd string, logFunc ...func(scanner *bufio.Scanner)) error {
// 	fmt.Println(cmd)
// 	in, err := s.sshSess.StdinPipe()
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	out, err := s.sshSess.StdoutPipe()
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	var output []byte
// 	r, w := io.Pipe()

// 	go func(in io.WriteCloser, out io.Reader, output *[]byte) {
// 		var (
// 			line string
// 			r    = bufio.NewReader(out)
// 		)
// 		for {
// 			b, err := r.ReadByte()
// 			if err != nil {
// 				break
// 			}
// 			*output = append(*output, b)
// 			if len(logFunc) > 0 {
// 				w.Write([]byte{b})
// 			}
// 			s.output.WriteByte(b)

// 			if b == byte('\n') {
// 				line = ""
// 				continue
// 			}

// 			line += string(b)

// 			if strings.HasPrefix(line, "[sudo] password for ") && strings.HasSuffix(line, ": ") {
// 				_, err = in.Write([]byte(s.sudoPasswd[0] + "\n"))
// 				if err != nil {
// 					break
// 				}
// 			}
// 		}
// 	}(in, out, &output)
// 	if len(logFunc) > 0 {
// 		go logFunc[0](bufio.NewScanner(r))
// 	}

// 	cmd = fmt.Sprintf(`sh -c  "%s"`, cmd)

// 	if len(s.sudoPasswd) > 0 {
// 		cmd = "sudo " + cmd
// 	}
// 	fmt.Println(cmd)

// 	err = s.sshSess.Start(cmd)
// 	s.output = bytes.NewBuffer(output)
// 	//wait stdout deal complete
// 	if len(logFunc) > 0 {
// 		time.Sleep(2 * time.Second)
// 	}
// 	return err
// }
